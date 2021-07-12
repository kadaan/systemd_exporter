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
		ret, err2 := fs.NewCPUStat(cgSubpath)
		if ret == nil {
			return nil, err2
		} else {
			return ret, nil
		}
	} else {
		ret, err2 := fs.NewCPUAcct(cgSubpath)
		if ret == nil {
			return nil, err2
		} else {
			return ret, nil
		}
	}
}
