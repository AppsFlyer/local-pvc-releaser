package objects

type Utils interface {
	PersistentVolumeClaim() PVC
	PersistentVolume() PV
	Event() Event
}

type utils struct {
}

func (_ *utils) PersistentVolumeClaim() PVC { return NewPVC() }
func (_ *utils) PersistentVolume() PV       { return NewPV() }
func (_ *utils) Event() Event               { return NewEvent() }

func Helper() Utils {
	return &utils{}
}
