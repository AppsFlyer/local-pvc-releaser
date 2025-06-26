package objects

type Utils interface {
	PersistentVolumeClaim() PVC
	PersistentVolume() PV
	Event() Event
}

type utils struct {
}

func (*utils) PersistentVolumeClaim() PVC { return NewPVC() }
func (*utils) PersistentVolume() PV       { return NewPV() }
func (*utils) Event() Event               { return NewEvent() }

func Helper() Utils {
	return &utils{}
}
