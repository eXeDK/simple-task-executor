package taskExecutor

import (
	"github.com/simple-task-executor/golang/target"
	"github.com/simple-task-executor/golang/taskHandler/taskExecutor/certCheck"
)

func ExecuteTask(targetConfig target.Config) interface{} {
	switch targetConfig.TargetType {
	case target.TypeCertCheck:
		return certCheck.Handle(targetConfig.Config)

	default:
		return nil
	}
}
