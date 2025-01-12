package coldmoon

import (
	"path"

	"github.com/Seeingu/coldmoon/pkg"
)

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
	agent.QueuedPromiseJobs.Push(&QueuedPromiseJob{
		job:   job,
		realm: realm,
	})
}

func HostPromiseRejectionTracker(promise *PromiseObject, operation PromiseRejectionTrackerOperation) {
	// TODO
}

func HostCallJobCallback(callback *JobCallback, this Value, arguments []Value) Value {
	Assert(IsCallable((callback.Callback).ToValue()))
	return (callback.Callback).ToValue().Call(this, arguments)
}

// MARK: - HostResizeArrayBuffer 25.1.3.8

type ResizeArrayBufferHandled int

const (
	ResizeArrayBufferHandledHandled ResizeArrayBufferHandled = iota
	ResizeArrayBufferHandledUnhandled
)

func HostResizeArrayBuffer(buffer *ArrayBufferLike, newByteLength JSInt) ResizeArrayBufferHandled {
	return ResizeArrayBufferHandledUnhandled
}

type ImportMetaProperties map[PropertyKey]Value

func HostGetImportMetaProperties(module *SourceTextModule) (i ImportMetaProperties) {
	return
}

func HostFinalizeImportMeta(meta ObjectType, module *SourceTextModule) {
	// return UNUSED
}

func HostLoadImportedModule(
	agent *Agent,
	referrer ImportedModuleReferrer,
	specifier string,
	hostDefined HostDefined,
	payload ImportedModulePayload,
) {
	filePath := resolveModulePath(specifier, hostDefined)
	sourceText := pkg.MustReadFile(filePath)
	module := ParseModule(sourceText, agent.CurrentRealm(), HostDefined{
		FileName: path.Base(filePath),
		BaseDir:  path.Dir(filePath),
	})
	result := NewCompletionModule(module)
	FinishLoadingImportedModule(agent, referrer, specifier, payload, result)
}

// TODO: handle protocols
func resolveModulePath(specifier string, hostDefined HostDefined) string {
	return path.Join(hostDefined.BaseDir, specifier)
}
