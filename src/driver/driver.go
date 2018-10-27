package main

/*
 * This class extends the volume driver helper class.
 */
import (
	"fmt"
	"log/syslog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/docker/go-plugins-helpers/volume"
)

type fileVolDriver struct {
	vsize  string
	vpath  string
	fstyp  string
	mhome  string
	mutex  sync.RWMutex
	logger *syslog.Writer
}

func newDriver(vsize, vpath, fstyp, mhome string) (*fileVolDriver, error) {
	logger, err := syslog.New(syslog.LOG_ERR, "docker-filevol-plugin")
	if err != nil {
		return nil, err
	}

	return &fileVolDriver{
		vsize:  vsize,
		vpath:  vpath,
		fstyp:  fstyp,
		mhome:  mhome,
		logger: logger,
	}, nil
}

func getVolumePath(vpath, fstyp, name string) string {
	s := []string{path.Join(vpath, name), "img"}
	return strings.Join(s, ".")
}

func getVolumeRegx(vpath, fstyp string) string {
	return path.Join(vpath, "*.img")
}

func getVolumeName(vpath string) string {
	return strings.TrimSuffix(path.Base(vpath), ".img")
}

func getMountHome(mhome, fstyp string) string {
	return mhome
}

func getMountPath(mhome, fstyp, name string) string {
	return path.Join(getMountHome(mhome, fstyp), name)
}

func (vd *fileVolDriver) Create(req *volume.CreateRequest) error {
	var (
		source  string
		volsize string
	)

	vd.mutex.Lock()
	defer vd.mutex.Unlock()

	// Prepare the path of the target volume.
	vpath := getVolumePath(vd.vpath, vd.fstyp, req.Name)

	// Do nothing if the target volume already exists.
	if _, err := os.Stat(vpath); !os.IsNotExist(err) {
		vd.logger.Err(fmt.Sprintf("Create: error: target volume exists. %s.", err))
		return err
	}

	// Fetch create volume command line options.
	for key, value := range req.Options {
		if key == "size" {
			volsize = value
		} else if key == "source" {
			source = value
		}
	}

	// Create snapshot volume, if requested.
	if source != "" {
		snapof := getVolumePath(vd.vpath, vd.fstyp, source)

		cmd := exec.Command("cp", "-np", snapof, vpath)
		if out, err := cmd.CombinedOutput(); err != nil {
			vd.logger.Err(fmt.Sprintf("Create: snapshot error: %s. %s.", err, string(out)))
			return err
		}
	} else {
		if volsize == "" {
			volsize = vd.vsize
		}

		ofvol := fmt.Sprintf("of=%s", vpath)
		count := fmt.Sprintf("count=%s", volsize)
		cmd1 := exec.Command("dd", "if=/dev/zero", ofvol, count)
		if out, err := cmd1.CombinedOutput(); err != nil {
			vd.logger.Err(fmt.Sprintf("Create: dd error: %s. %s.", err, string(out)))
			return err
		}

		cmd2 := exec.Command("mkfs", "-t", vd.fstyp, "-F", vpath)
		if out, err := cmd2.CombinedOutput(); err != nil {
			exec.Command("rm", "-f", vpath)
			vd.logger.Err(fmt.Sprintf("Create: mkfs error: %s. %s.", err, string(out)))
			return err
		}
	}

	return nil
}

func (vd *fileVolDriver) Remove(req *volume.RemoveRequest) error {
	vpath := getVolumePath(vd.vpath, vd.fstyp, req.Name)
	if _, err := os.Stat(vpath); os.IsNotExist(err) {
		return nil
	}

	cmd := exec.Command("rm", "-f", vpath)
	if out, err := cmd.CombinedOutput(); err != nil {
		vd.logger.Err(fmt.Sprintf("Remove: rm error: %s. %s.", err, string(out)))
		return err
	}

	return nil
}

func (vd *fileVolDriver) Path(req *volume.PathRequest) (*volume.PathResponse, error) {
	mount := getMountPath(vd.mhome, vd.fstyp, req.Name)
	return &volume.PathResponse{Mountpoint: mount}, nil
}

func (vd *fileVolDriver) Mount(req *volume.MountRequest) (*volume.MountResponse, error) {
	vd.mutex.Lock()
	defer vd.mutex.Unlock()

	mount := getMountPath(vd.mhome, vd.fstyp, req.Name)
	vpath := getVolumePath(vd.vpath, vd.fstyp, req.Name)

	err1 := os.MkdirAll(mount, 0700)
	if err1 != nil {
		vd.logger.Err(fmt.Sprintf("Mount: mkdir error: %s.", err1))
		return &volume.MountResponse{Mountpoint: ""}, err1
	}

	cmd := exec.Command("mount", "-o", "loop", "-t", vd.fstyp, vpath, mount)
	out, err2 := cmd.CombinedOutput()
	if err2 != nil {
		vd.logger.Err(fmt.Sprintf("Mount: mount error: %s. %s.", err2, string(out)))
	}

	return &volume.MountResponse{Mountpoint: mount}, err2
}

func (vd *fileVolDriver) Unmount(req *volume.UnmountRequest) error {
	vd.mutex.Lock()
	defer vd.mutex.Unlock()

	mount := getMountPath(vd.mhome, vd.fstyp, req.Name)

	cmd := exec.Command("umount", mount)
	out, err := cmd.CombinedOutput()
	if err != nil {
		vd.logger.Err(fmt.Sprintf("Unmount: umount error: %s. %s.", err, string(out)))
	}

	return err
}

func (vd *fileVolDriver) Get(req *volume.GetRequest) (*volume.GetResponse, error) {
	vd.mutex.RLock()
	defer vd.mutex.RUnlock()

	vpath := getVolumePath(vd.vpath, vd.fstyp, req.Name)
	_, err := os.Stat(vpath)

	mount := getMountPath(vd.mhome, vd.fstyp, req.Name)

	vol := &volume.Volume{Name: req.Name, Mountpoint: mount}
	res := &volume.GetResponse{Volume: vol}

	return res, err
}

func (vd *fileVolDriver) List() (*volume.ListResponse, error) {
	vd.mutex.RLock()
	defer vd.mutex.RUnlock()

	var vols []*volume.Volume

	vfiles, err := filepath.Glob(getVolumeRegx(vd.vpath, vd.fstyp))

	for _, vfile := range vfiles {
		name := getVolumeName(vfile)
		mount := getMountPath(vd.mhome, vd.fstyp, name)
		vol := &volume.Volume{Name: name, Mountpoint: mount}
		vols = append(vols, vol)
	}

	return &volume.ListResponse{Volumes: vols}, err
}

func (vd *fileVolDriver) Capabilities() *volume.CapabilitiesResponse {
	capability := volume.Capability{Scope: "local"}
	return &volume.CapabilitiesResponse{Capabilities: capability}
}
