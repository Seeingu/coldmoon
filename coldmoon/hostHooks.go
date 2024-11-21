package coldmoon

// 19.2.1.2
func HostEnsureCanCompileStrings(realm *Realm) {

}

// 20.2.5
func HostHasSourceTextAvailable(o ObjectType) bool {
	return true
}

// 9.5.2
func HostMakeJobCallback(callback ObjectType) *JobCallback {
	return &JobCallback{
		Callback:    callback,
		HostDefined: nil,
	}
}

type PromiseRejectionTrackerOperation int

const (
	PromiseRejectionTrackerOperationReject PromiseRejectionTrackerOperation = iota
	PromiseRejectionTrackerOperationHandle
)

// 9.5.5
func HostEnqueuePromiseJob(agent *Agent, job *Job, realm *Realm) {
	// TODO
}

func HostPromiseRejectionTracker(promise *PromiseObject, operation PromiseRejectionTrackerOperation) {
	// TODO
}
