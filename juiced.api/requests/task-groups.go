package requests

import (
	"fmt"

	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	"backend.juicedbot.io/juiced.infrastructure/stores"
	"backend.juicedbot.io/juiced.sitescripts/pokemoncenter"
)

type CreateTaskGroupRequest struct {
	Name     string                  `json:"name"`
	Retailer string                  `json:"retailer"`
	Monitors []*entities.BaseMonitor `json:"monitors"`
}

type TaskGroupExistsError struct {
	Name string
}

func (e *TaskGroupExistsError) Error() string {
	return fmt.Sprintf("a TaskGroup with the name %s already exists", e.Name)
}

type NegativeDelayError struct{}

func (e *NegativeDelayError) Error() string {
	return "monitor delay cannot be negative"
}

type NegativeMaxPriceError struct{}

func (e *NegativeMaxPriceError) Error() string {
	return "max price cannot be negative"
}

type EmptyFieldError struct {
	Field string
}

func (e *EmptyFieldError) Error() string {
	return fmt.Sprintf("%s cannot be empty", e.Field)
}

func ValidateCreateTaskGroupRequest(request *CreateTaskGroupRequest) error {
	// Validate Name
	taskGroups := stores.GetAllTaskGroups()
	for _, taskGroup := range taskGroups {
		if taskGroup.Name == request.Name {
			return &TaskGroupExistsError{request.Name}
		}
	}

	// Validate Retailer
	if request.Retailer == "" {
		return &EmptyFieldError{"retailer"}
	}
	if err := enums.IsValidRetailer(request.Retailer); err != nil {
		return err
	}

	// Validate each SingleMonitorInput
	for _, monitor := range request.Monitors {
		// Validate ProxyGroup
		if monitor.MonitorInput.ProxyGroupID != "" {
			if _, err := stores.GetProxyGroup(monitor.MonitorInput.ProxyGroupID); err != nil {
				return err
			}
		}

		// Validate DelayMS
		if monitor.MonitorInput.DelayMS < 0 {
			return &NegativeDelayError{}
		}
		if monitor.MonitorInput.DelayMS == 0 {
			monitor.MonitorInput.DelayMS = 2000
		}

		// Validate MaxPrice
		if monitor.MonitorInput.MaxPrice < 0 {
			return &NegativeMaxPriceError{}
		}

		// Validate Input
		if monitor.MonitorInput.Input == "" {
			return &EmptyFieldError{"monitor input"}
		}

		// Validate MonitorType
		if monitor.MonitorInput.MonitorType == "" {
			return &EmptyFieldError{"monitor type"}
		}
		if err := enums.IsValidMonitorType(monitor.MonitorInput.MonitorType, request.Retailer); err != nil {
			return err
		}

		// Validate SiteSpecificInfo
		switch request.Retailer {
		case enums.PokemonCenter:
			if err := pokemoncenter.ValidateMonitorInput(monitor.MonitorInput.Input, monitor.MonitorInput.MonitorType, monitor.MonitorInput.SiteSpecificInfo); err != nil {
				return err
			}
		}
	}

	return nil
}
