package permission

const (
	None             = uint64(0b0) << iota
	RunOwnContainer  // create own container
	ManageContainers // manage all container
)
