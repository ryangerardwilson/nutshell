package compiler

import (
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

func nativeArtifact(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() == 0 {
		return fmt.Errorf("artifact must be a nonempty regular native executable")
	}
	if runtime.GOOS != "windows" && info.Mode()&0111 == 0 {
		return fmt.Errorf("artifact is not executable")
	}
	switch runtime.GOOS {
	case "linux":
		f, err := elf.Open(path)
		if err != nil {
			return fmt.Errorf("expected native ELF executable: %w", err)
		}
		defer f.Close()
		machine := map[string]elf.Machine{"amd64": elf.EM_X86_64, "arm64": elf.EM_AARCH64, "386": elf.EM_386, "arm": elf.EM_ARM, "riscv64": elf.EM_RISCV, "ppc64le": elf.EM_PPC64, "s390x": elf.EM_S390}[runtime.GOARCH]
		if machine == 0 || f.Machine != machine || (f.Type != elf.ET_EXEC && f.Type != elf.ET_DYN) || f.Entry == 0 {
			return fmt.Errorf("artifact is not a %s executable", runtime.GOARCH)
		}
	case "darwin":
		f, err := macho.Open(path)
		if err != nil {
			return fmt.Errorf("expected native Mach-O executable: %w", err)
		}
		defer f.Close()
		cpu := map[string]macho.Cpu{"amd64": macho.CpuAmd64, "arm64": macho.CpuArm64}[runtime.GOARCH]
		if f.Type != macho.TypeExec || cpu == 0 || f.Cpu != cpu {
			return fmt.Errorf("artifact is not a %s executable", runtime.GOARCH)
		}
	case "windows":
		f, err := pe.Open(path)
		if err != nil {
			return fmt.Errorf("expected native PE executable: %w", err)
		}
		defer f.Close()
		machine := map[string]uint16{"amd64": pe.IMAGE_FILE_MACHINE_AMD64, "arm64": pe.IMAGE_FILE_MACHINE_ARM64, "386": pe.IMAGE_FILE_MACHINE_I386}[runtime.GOARCH]
		if machine == 0 || f.Machine != machine || f.Characteristics&pe.IMAGE_FILE_DLL != 0 || f.Characteristics&pe.IMAGE_FILE_EXECUTABLE_IMAGE == 0 {
			return fmt.Errorf("artifact is not a %s executable", runtime.GOARCH)
		}
	default:
		return fmt.Errorf("native validation is not supported on %s", runtime.GOOS)
	}
	return nil
}

func outputPath(p Program, requested string) (string, error) {
	if requested == "" {
		requested = filepath.Join(p.Root, p.Entry[:len(p.Entry)-len(".nut")])
		if runtime.GOOS == "windows" {
			requested += ".exe"
		}
	}
	abs, err := filepath.Abs(requested)
	if err != nil {
		return "", err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(abs))
	if err != nil {
		return "", fmt.Errorf("output directory must exist: %w", err)
	}
	abs = filepath.Join(parent, filepath.Base(abs))
	if filepath.Ext(abs) == ".nut" {
		return "", fmt.Errorf("cannot overwrite .nut source: %s", abs)
	}
	if rel, err := filepath.Rel(filepath.Join(p.Root, ".nutshell"), abs); err == nil && filepath.IsLocal(rel) {
		return "", fmt.Errorf("output must be outside .nutshell build metadata")
	}
	for _, src := range p.Sources {
		if abs == filepath.Join(p.Root, filepath.FromSlash(src.Path)) {
			return "", fmt.Errorf("output would overwrite source: %s", abs)
		}
	}
	if _, err := os.Lstat(abs); err == nil {
		if err := nativeArtifact(abs); err != nil {
			return "", fmt.Errorf("refusing to replace a non-native output %s: %w", abs, err)
		}
	} else if !os.IsNotExist(err) {
		return "", err
	}
	return abs, nil
}

func publish(artifact, destination string) error {
	src, err := os.Open(artifact)
	if err != nil {
		return err
	}
	defer src.Close()
	tmp, err := os.CreateTemp(filepath.Dir(destination), ".nutshell-output-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()
	if _, err = io.Copy(tmp, src); err != nil {
		return err
	}
	if err = tmp.Chmod(0755); err != nil {
		return err
	}
	if err = tmp.Sync(); err != nil {
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	if err = nativeArtifact(tmp.Name()); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), destination)
}
