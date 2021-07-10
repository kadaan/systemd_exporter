package cgroup

type TotalCPUUsage interface {
	SystemSeconds() float64
	UserSeconds() float64
}

func NewCPUUsage(cgSubpath string) (TotalCPUUsage, error) {
	fs, err := NewDefaultFS()
	if err != nil {
		return nil, err
	}

	if fs.cgroupUnified == MountModeUnified || fs.cgroupUnified == MountModeHybrid {
		return fs.NewCPUStat(cgSubpath)
	} else {
		return fs.NewCPUAcct(cgSubpath)
	}
}
