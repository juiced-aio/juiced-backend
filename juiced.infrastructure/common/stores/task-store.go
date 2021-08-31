package stores

import (
	e "errors"
	"strings"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/errors"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.infrastructure/queries"

	"backend.juicedbot.io/juiced.sitescripts/amazon"
	"backend.juicedbot.io/juiced.sitescripts/bestbuy"
	"backend.juicedbot.io/juiced.sitescripts/boxlunch"
	"backend.juicedbot.io/juiced.sitescripts/disney"
	"backend.juicedbot.io/juiced.sitescripts/gamestop"
	"backend.juicedbot.io/juiced.sitescripts/hottopic"
	"backend.juicedbot.io/juiced.sitescripts/newegg"
	"backend.juicedbot.io/juiced.sitescripts/pokemoncenter"
	"backend.juicedbot.io/juiced.sitescripts/shopify"
	"backend.juicedbot.io/juiced.sitescripts/target"
	"backend.juicedbot.io/juiced.sitescripts/topps"
	"backend.juicedbot.io/juiced.sitescripts/walmart"
	// Future sitescripts will be imported here
)

// TODO @silent: Handle TaskUpdatedEvent by updating the task's info in here
// TODO @silent: Handle TaskRemovedEvent by removing the task's info in here
// TODO @silent: Handle TaskGroupRemovedEvent by removing all of the task group's tasks in here

// TaskStore stores information about running Tasks
type TaskStore struct {
	AmazonTasks        map[string]*amazon.Task
	BestbuyTasks       map[string]*bestbuy.Task
	BoxlunchTasks      map[string]*boxlunch.Task
	DisneyTasks        map[string]*disney.Task
	GamestopTasks      map[string]*gamestop.Task
	HottopicTasks      map[string]*hottopic.Task
	NeweggTasks        map[string]*newegg.Task
	PokemonCenterTasks map[string]*pokemoncenter.Task
	ShopifyTasks       map[string]*shopify.Task
	TargetTasks        map[string]*target.Task
	ToppsTasks         map[string]*topps.Task
	WalmartTasks       map[string]*walmart.Task

	// Future sitescripts will have a field here
	EventBus *events.EventBus
}

// AddTaskToStore adds the Task to the TaskStore and returns true if successful
func (taskStore *TaskStore) AddTaskToStore(task *entities.Task) error {
	var queryError error
	var profile entities.Profile
	// Get Profile, ProxyGroup for task
	if task.TaskRetailer != enums.Amazon {
		var err error
		profile, err = queries.GetProfile(task.TaskProfileID)
		if err != nil {
			queryError = err
		}
	}
	var proxyGroup *entities.ProxyGroup
	if task.TaskProxyGroupID != "" {
		var ok bool
		proxyGroup, ok = proxyStore.ProxyGroups[task.TaskProxyGroupID]
		if !ok {
			queryError = e.New("proxy group failure")
		}
	}
	switch task.TaskRetailer {
	// Future sitescripts will have a case here
	case enums.Amazon:
		// Check if task exists in store already
		if _, ok := taskStore.AmazonTasks[task.ID]; ok && !task.UpdateTask {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}
		// Make sure necessary fields exist
		emptyString := ""
		if task.AmazonTaskInfo.Email == emptyString || task.AmazonTaskInfo.Password == emptyString {
			return e.New(errors.MissingTaskFieldsError)
		}
		if task.AmazonTaskInfo.LoginType == emptyString {
			task.AmazonTaskInfo.LoginType = enums.LoginTypeBROWSER
		}

		// Create task
		amazonTask, err := amazon.CreateAmazonTask(task, proxyGroup, taskStore.EventBus, task.AmazonTaskInfo.LoginType, task.AmazonTaskInfo.Email, task.AmazonTaskInfo.Password)
		if err != nil {
			return e.New(errors.CreateBotTaskError + err.Error())
		}
		// Add task to store
		taskStore.AmazonTasks[task.ID] = &amazonTask

	case enums.BestBuy:
		// Check if task exists in store already
		if _, ok := taskStore.BestbuyTasks[task.ID]; ok && !task.UpdateTask {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}
		// Make sure necessary fields exist
		emptyString := ""
		if task.BestbuyTaskInfo.TaskType == emptyString || (task.BestbuyTaskInfo.TaskType == enums.TaskTypeAccount && (task.BestbuyTaskInfo.Email == emptyString || task.BestbuyTaskInfo.Password == emptyString)) {
			return e.New(errors.MissingTaskFieldsError)
		}
		// Create task
		bestbuyTask, err := bestbuy.CreateBestbuyTask(task, profile, proxyGroup, taskStore.EventBus, task.BestbuyTaskInfo.TaskType, task.BestbuyTaskInfo.LocationID, task.BestbuyTaskInfo.Email, task.BestbuyTaskInfo.Password)
		if err != nil {
			return e.New(errors.CreateBotTaskError + err.Error())
		}
		// Add task to store
		taskStore.BestbuyTasks[task.ID] = &bestbuyTask

	case enums.BoxLunch:
		// Check if task exists in store already
		if _, ok := taskStore.BoxlunchTasks[task.ID]; ok && !task.UpdateTask {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}
		// Create task
		boxlunchTask, err := boxlunch.CreateBoxlunchTask(task, profile, proxyGroup, taskStore.EventBus)
		if err != nil {
			return e.New(errors.CreateBotTaskError + err.Error())
		}
		// Add task to store
		taskStore.BoxlunchTasks[task.ID] = &boxlunchTask

	case enums.GameStop:
		// Check if task exists in store already
		if _, ok := taskStore.GamestopTasks[task.ID]; ok && !task.UpdateTask {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}

		// Make sure necessary fields exist
		emptyString := ""
		if task.GamestopTaskInfo.TaskType == emptyString || (task.GamestopTaskInfo.TaskType == enums.TaskTypeAccount && (task.GamestopTaskInfo.Email == emptyString || task.GamestopTaskInfo.Password == emptyString)) {
			return e.New(errors.MissingTaskFieldsError)
		}

		// Create task
		gamestopTask, err := gamestop.CreateGamestopTask(task, profile, proxyGroup, taskStore.EventBus, task.GamestopTaskInfo.TaskType, task.GamestopTaskInfo.Email, task.GamestopTaskInfo.Password)
		if err != nil {
			return e.New(errors.CreateBotTaskError + err.Error())
		}
		// Add task to store
		taskStore.GamestopTasks[task.ID] = &gamestopTask

	case enums.Disney:
		// Check if task exists in store already
		if _, ok := taskStore.DisneyTasks[task.ID]; ok && !task.UpdateTask {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}
		// Make sure necessary fields exist
		emptyString := ""
		if task.DisneyTaskInfo.TaskType == emptyString || (task.DisneyTaskInfo.TaskType == enums.TaskTypeAccount && (task.DisneyTaskInfo.Email == emptyString || task.DisneyTaskInfo.Password == emptyString)) {
			return e.New(errors.MissingTaskFieldsError)
		}
		// Create task
		disneyTask, err := disney.CreateDisneyTask(task, profile, proxyGroup, taskStore.EventBus, task.DisneyTaskInfo.TaskType, task.DisneyTaskInfo.Email, task.DisneyTaskInfo.Password)
		if err != nil {
			return e.New(errors.CreateBotTaskError + err.Error())
		}
		// Add task to store
		taskStore.DisneyTasks[task.ID] = &disneyTask

	case enums.HotTopic:
		// Check if task exists in store already
		if _, ok := taskStore.HottopicTasks[task.ID]; ok && !task.UpdateTask {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}
		// Create task
		hottopicTask, err := hottopic.CreateHottopicTask(task, profile, proxyGroup, taskStore.EventBus)
		if err != nil {
			return e.New(errors.CreateBotTaskError + err.Error())
		}
		// Add task to store
		taskStore.HottopicTasks[task.ID] = &hottopicTask

	case enums.Newegg:
		// Check if task exists in store already
		if _, ok := taskStore.NeweggTasks[task.ID]; ok && !task.UpdateTask {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}
		// Create task
		neweggTask, err := newegg.CreateNeweggTask(task, profile, proxyGroup, taskStore.EventBus)
		if err != nil {
			return e.New(errors.CreateBotTaskError + err.Error())
		}
		// Add task to store
		taskStore.NeweggTasks[task.ID] = &neweggTask

	case enums.PokemonCenter:
		// Check if task exists in store already
		if _, ok := taskStore.PokemonCenterTasks[task.ID]; ok {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}
		// Make sure necessary fields exist
		emptyString := ""
		if task.PokemonCenterTaskInfo.TaskType == emptyString || task.PokemonCenterTaskInfo.AddressType == emptyString || (task.PokemonCenterTaskInfo.TaskType == enums.TaskTypeAccount && (task.PokemonCenterTaskInfo.Email == emptyString || task.PokemonCenterTaskInfo.Password == emptyString)) {
			return e.New(errors.MissingTaskFieldsError)
		}
		// Create task
		pokemonCenterTask, err := pokemoncenter.CreatePokemonCenterTask(task, profile, proxyGroup, taskStore.EventBus, task.PokemonCenterTaskInfo.Email, task.PokemonCenterTaskInfo.Password, task.PokemonCenterTaskInfo.TaskType)
		if err != nil {
			return e.New(errors.CreateBotTaskError + err.Error())
		}
		// Add task to store
		taskStore.PokemonCenterTasks[task.ID] = &pokemonCenterTask

	case enums.Shopify:
		// Check if task exists in store already
		if _, ok := taskStore.ShopifyTasks[task.ID]; ok && !task.UpdateTask {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}
		// Make sure necessary fields exist
		emptyString := ""
		if task.ShopifyTaskInfo.SiteURL == emptyString || task.ShopifyTaskInfo.ShopifyRetailer == emptyString {
			return e.New(errors.MissingTaskFieldsError)
		}

		// Shopify Site specifics
		site := task.ShopifyTaskInfo.ShopifyRetailer
		switch site {
		case enums.HotWheels:
			if task.ShopifyTaskInfo.HotWheelsTaskInfo.Email != "" && task.ShopifyTaskInfo.HotWheelsTaskInfo.Password != "" {
				return e.New(errors.MissingTaskFieldsError)
			}
		}
		// Create task
		shopifyTask, err := shopify.CreateShopifyTask(task, profile, proxyGroup, taskStore.EventBus, task.ShopifyTaskInfo.CouponCode, task.ShopifyTaskInfo.SiteURL, task.ShopifyTaskInfo.SitePassword, task.ShopifyTaskInfo.HotWheelsTaskInfo.Email, task.ShopifyTaskInfo.HotWheelsTaskInfo.Password)
		if err != nil {
			return e.New(errors.CreateBotTaskError + err.Error())
		}
		// Add task to store
		taskStore.ShopifyTasks[task.ID] = &shopifyTask

	case enums.Target:
		// Check if task exists in store already
		if _, ok := taskStore.TargetTasks[task.ID]; ok && !task.UpdateTask {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}
		// Make sure necessary fields exist
		emptyString := ""
		if task.TargetTaskInfo.Email == emptyString || task.TargetTaskInfo.Password == emptyString || task.TargetTaskInfo.PaymentType == emptyString {
			return e.New(errors.MissingTaskFieldsError)
		}
		// Create task
		targetTask, err := target.CreateTargetTask(task, profile, proxyGroup, taskStore.EventBus, task.TargetTaskInfo.Email, task.TargetTaskInfo.Password, task.TargetTaskInfo.PaymentType)
		if err != nil {
			return e.New(errors.CreateBotTaskError + err.Error())
		}
		// Add task to store
		taskStore.TargetTasks[task.ID] = &targetTask

	case enums.Topps:
		// Check if task exists in store already
		if _, ok := taskStore.ToppsTasks[task.ID]; ok && !task.UpdateTask {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}

		// Make sure necessary fields exist
		emptyString := ""
		if task.ToppsTaskInfo.TaskType == emptyString || (task.ToppsTaskInfo.TaskType == enums.TaskTypeAccount && (task.ToppsTaskInfo.Email == emptyString || task.ToppsTaskInfo.Password == emptyString)) {
			return e.New(errors.MissingTaskFieldsError)
		}

		// Create task
		toppsTask, err := topps.CreateToppsTask(task, profile, proxyGroup, taskStore.EventBus, task.ToppsTaskInfo.TaskType, task.ToppsTaskInfo.Email, task.ToppsTaskInfo.Password)
		if err != nil {
			return e.New(errors.CreateBotTaskError + err.Error())
		}
		// Add task to store
		taskStore.ToppsTasks[task.ID] = &toppsTask

	case enums.Walmart:
		// Check if task exists in store already
		if _, ok := taskStore.WalmartTasks[task.ID]; ok && !task.UpdateTask {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}

		// Make sure necessary fields exist
		emptyString := ""
		if task.WalmartTaskInfo.TaskType == emptyString || (task.WalmartTaskInfo.TaskType == enums.TaskTypeAccount && (task.WalmartTaskInfo.Email == emptyString || task.WalmartTaskInfo.Password == emptyString)) {
			return e.New(errors.MissingTaskFieldsError)
		}

		// Create task
		walmartTask, err := walmart.CreateWalmartTask(task, profile, proxyGroup, taskStore.EventBus, task.WalmartTaskInfo.Email, task.WalmartTaskInfo.Password, task.WalmartTaskInfo.TaskType)
		if err != nil {
			return e.New(errors.CreateBotTaskError + err.Error())
		}
		// Add task to store
		taskStore.WalmartTasks[task.ID] = &walmartTask

	}
	task.UpdateTask = false
	return nil
}

// AddTestTaskToStore adds the Test Task to the TaskStore and returns true if successful
func (taskStore *TaskStore) AddTestTaskToStore(task *entities.Task, profile entities.Profile) error {
	var queryError error
	// Get Profile, Proxy for task
	var proxyGroup *entities.ProxyGroup
	if task.TaskProxyGroupID != "" {
		var ok bool
		proxyGroup, ok = proxyStore.ProxyGroups[task.TaskProxyGroupID]
		if !ok {
			queryError = e.New("proxy group failure")
		}
	}

	switch task.TaskRetailer {
	// Future sitescripts will have a case here
	case enums.Amazon:
		// Check if task exists in store already
		if _, ok := taskStore.AmazonTasks[task.ID]; ok && !task.UpdateTask {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}
		// Make sure necessary fields exist
		emptyString := ""
		if task.AmazonTaskInfo.Email == emptyString || task.AmazonTaskInfo.Password == emptyString {
			return e.New(errors.MissingTaskFieldsError)
		}
		if task.AmazonTaskInfo.LoginType == emptyString {
			task.AmazonTaskInfo.LoginType = enums.LoginTypeBROWSER
		}

		// Create task
		amazonTask, err := amazon.CreateAmazonTask(task, proxyGroup, taskStore.EventBus, task.AmazonTaskInfo.LoginType, task.AmazonTaskInfo.Email, task.AmazonTaskInfo.Password)
		if err != nil {
			return e.New(errors.CreateBotTaskError + err.Error())
		}
		// Add task to store
		taskStore.AmazonTasks[task.ID] = &amazonTask

	case enums.BestBuy:
		// Check if task exists in store already
		if _, ok := taskStore.BestbuyTasks[task.ID]; ok && !task.UpdateTask {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}
		// Make sure necessary fields exist
		emptyString := ""
		if task.BestbuyTaskInfo.TaskType == emptyString || (task.BestbuyTaskInfo.TaskType == enums.TaskTypeAccount && (task.BestbuyTaskInfo.Email == emptyString || task.BestbuyTaskInfo.Password == emptyString)) {
			return e.New(errors.MissingTaskFieldsError)
		}
		// Create task
		bestbuyTask, err := bestbuy.CreateBestbuyTask(task, profile, proxyGroup, taskStore.EventBus, task.BestbuyTaskInfo.TaskType, task.BestbuyTaskInfo.LocationID, task.BestbuyTaskInfo.Email, task.BestbuyTaskInfo.Password)
		if err != nil {
			return e.New(errors.CreateBotTaskError + err.Error())
		}
		// Add task to store
		taskStore.BestbuyTasks[task.ID] = &bestbuyTask

	case enums.BoxLunch:
		// Check if task exists in store already
		if _, ok := taskStore.BoxlunchTasks[task.ID]; ok && !task.UpdateTask {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}
		// Create task
		boxlunchTask, err := boxlunch.CreateBoxlunchTask(task, profile, proxyGroup, taskStore.EventBus)
		if err != nil {
			return e.New(errors.CreateBotTaskError + err.Error())
		}
		// Add task to store
		taskStore.BoxlunchTasks[task.ID] = &boxlunchTask

	case enums.GameStop:
		// Check if task exists in store already
		if _, ok := taskStore.GamestopTasks[task.ID]; ok && !task.UpdateTask {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}

		// Make sure necessary fields exist
		emptyString := ""
		if task.GamestopTaskInfo.TaskType == emptyString || (task.GamestopTaskInfo.TaskType == enums.TaskTypeAccount && (task.GamestopTaskInfo.Email == emptyString || task.GamestopTaskInfo.Password == emptyString)) {
			return e.New(errors.MissingTaskFieldsError)
		}

		// Create task
		gamestopTask, err := gamestop.CreateGamestopTask(task, profile, proxyGroup, taskStore.EventBus, task.GamestopTaskInfo.TaskType, task.GamestopTaskInfo.Email, task.GamestopTaskInfo.Password)
		if err != nil {
			return e.New(errors.CreateBotTaskError + err.Error())
		}
		// Add task to store
		taskStore.GamestopTasks[task.ID] = &gamestopTask

	case enums.Disney:
		// Check if task exists in store already
		if _, ok := taskStore.DisneyTasks[task.ID]; ok && !task.UpdateTask {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}
		// Make sure necessary fields exist
		emptyString := ""
		if task.DisneyTaskInfo.TaskType == emptyString || (task.DisneyTaskInfo.TaskType == enums.TaskTypeAccount && (task.DisneyTaskInfo.Email == emptyString || task.DisneyTaskInfo.Password == emptyString)) {
			return e.New(errors.MissingTaskFieldsError)
		}
		// Create task
		disneyTask, err := disney.CreateDisneyTask(task, profile, proxyGroup, taskStore.EventBus, task.DisneyTaskInfo.TaskType, task.DisneyTaskInfo.Email, task.DisneyTaskInfo.Password)
		if err != nil {
			return e.New(errors.CreateBotTaskError + err.Error())
		}
		// Add task to store
		taskStore.DisneyTasks[task.ID] = &disneyTask

	case enums.HotTopic:
		// Check if task exists in store already
		if _, ok := taskStore.HottopicTasks[task.ID]; ok && !task.UpdateTask {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}
		// Create task
		hottopicTask, err := hottopic.CreateHottopicTask(task, profile, proxyGroup, taskStore.EventBus)
		if err != nil {
			return e.New(errors.CreateBotTaskError + err.Error())
		}
		// Add task to store
		taskStore.HottopicTasks[task.ID] = &hottopicTask

	case enums.Newegg:
		// Check if task exists in store already
		if _, ok := taskStore.NeweggTasks[task.ID]; ok && !task.UpdateTask {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}
		// Create task
		neweggTask, err := newegg.CreateNeweggTask(task, profile, proxyGroup, taskStore.EventBus)
		if err != nil {
			return e.New(errors.CreateBotTaskError + err.Error())
		}
		// Add task to store
		taskStore.NeweggTasks[task.ID] = &neweggTask

	case enums.PokemonCenter:
		// Check if task exists in store already
		if _, ok := taskStore.PokemonCenterTasks[task.ID]; ok {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}
		// Make sure necessary fields exist
		emptyString := ""
		if task.PokemonCenterTaskInfo.TaskType == emptyString || task.PokemonCenterTaskInfo.AddressType == emptyString || (task.PokemonCenterTaskInfo.TaskType == enums.TaskTypeAccount && (task.PokemonCenterTaskInfo.Email == emptyString || task.PokemonCenterTaskInfo.Password == emptyString)) {
			return e.New(errors.MissingTaskFieldsError)
		}
		// Create task
		pokemonCenterTask, err := pokemoncenter.CreatePokemonCenterTask(task, profile, proxyGroup, taskStore.EventBus, task.PokemonCenterTaskInfo.Email, task.PokemonCenterTaskInfo.Password, task.PokemonCenterTaskInfo.TaskType)
		if err != nil {
			return e.New(errors.CreateBotTaskError + err.Error())
		}
		// Add task to store
		taskStore.PokemonCenterTasks[task.ID] = &pokemonCenterTask

	case enums.Shopify:
		// Check if task exists in store already
		if _, ok := taskStore.ShopifyTasks[task.ID]; ok && !task.UpdateTask {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}
		// Make sure necessary fields exist
		emptyString := ""
		if task.ShopifyTaskInfo.SiteURL == emptyString || task.ShopifyTaskInfo.ShopifyRetailer == emptyString {
			return e.New(errors.MissingTaskFieldsError)
		}

		// Shopify Site specifics
		site := task.ShopifyTaskInfo.ShopifyRetailer
		switch site {
		case enums.HotWheels:
			if task.ShopifyTaskInfo.HotWheelsTaskInfo.Email != "" && task.ShopifyTaskInfo.HotWheelsTaskInfo.Password != "" {
				return e.New(errors.MissingTaskFieldsError)
			}
		}
		// Create task
		shopifyTask, err := shopify.CreateShopifyTask(task, profile, proxyGroup, taskStore.EventBus, task.ShopifyTaskInfo.CouponCode, task.ShopifyTaskInfo.SiteURL, task.ShopifyTaskInfo.SitePassword, task.ShopifyTaskInfo.HotWheelsTaskInfo.Email, task.ShopifyTaskInfo.HotWheelsTaskInfo.Password)
		if err != nil {
			return e.New(errors.CreateBotTaskError + err.Error())
		}
		// Add task to store
		taskStore.ShopifyTasks[task.ID] = &shopifyTask

	case enums.Target:
		// Check if task exists in store already
		if _, ok := taskStore.TargetTasks[task.ID]; ok && !task.UpdateTask {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}
		// Make sure necessary fields exist
		emptyString := ""
		if task.TargetTaskInfo.Email == emptyString || task.TargetTaskInfo.Password == emptyString || task.TargetTaskInfo.PaymentType == emptyString {
			return e.New(errors.MissingTaskFieldsError)
		}
		// Create task
		targetTask, err := target.CreateTargetTask(task, profile, proxyGroup, taskStore.EventBus, task.TargetTaskInfo.Email, task.TargetTaskInfo.Password, task.TargetTaskInfo.PaymentType)
		if err != nil {
			return e.New(errors.CreateBotTaskError + err.Error())
		}
		// Add task to store
		taskStore.TargetTasks[task.ID] = &targetTask

	case enums.Topps:
		// Check if task exists in store already
		if _, ok := taskStore.ToppsTasks[task.ID]; ok && !task.UpdateTask {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}

		// Make sure necessary fields exist
		emptyString := ""
		if task.ToppsTaskInfo.TaskType == emptyString || (task.ToppsTaskInfo.TaskType == enums.TaskTypeAccount && (task.ToppsTaskInfo.Email == emptyString || task.ToppsTaskInfo.Password == emptyString)) {
			return e.New(errors.MissingTaskFieldsError)
		}

		// Create task
		toppsTask, err := topps.CreateToppsTask(task, profile, proxyGroup, taskStore.EventBus, task.ToppsTaskInfo.TaskType, task.ToppsTaskInfo.Email, task.ToppsTaskInfo.Password)
		if err != nil {
			return e.New(errors.CreateBotTaskError + err.Error())
		}
		// Add task to store
		taskStore.ToppsTasks[task.ID] = &toppsTask

	case enums.Walmart:
		// Check if task exists in store already
		if _, ok := taskStore.WalmartTasks[task.ID]; ok && !task.UpdateTask {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}

		// Make sure necessary fields exist
		emptyString := ""
		if task.WalmartTaskInfo.TaskType == emptyString || (task.WalmartTaskInfo.TaskType == enums.TaskTypeAccount && (task.WalmartTaskInfo.Email == emptyString || task.WalmartTaskInfo.Password == emptyString)) {
			return e.New(errors.MissingTaskFieldsError)
		}

		// Create task
		walmartTask, err := walmart.CreateWalmartTask(task, profile, proxyGroup, taskStore.EventBus, task.WalmartTaskInfo.Email, task.WalmartTaskInfo.Password, task.WalmartTaskInfo.TaskType)
		if err != nil {
			return e.New(errors.CreateBotTaskError + err.Error())
		}
		// Add task to store
		taskStore.WalmartTasks[task.ID] = &walmartTask

	}
	task.UpdateTask = false
	return nil
}

// StartTaskGroup runs the given TaskGroup's RunMonitor() function and the RunTask() function for each Task in the group and returns true if successful
func (taskStore *TaskStore) StartTaskGroup(taskGroup *entities.TaskGroup) ([]string, error) {
	// Start the task's TaskGroup (if it's already running, this will return true)
	var warnings []string
	err := monitorStore.StartMonitor(taskGroup)
	if err != nil {
		return warnings, err
	}

	for _, taskID := range taskGroup.TaskIDs {
		// Get the task
		task, err := queries.GetTask(taskID)
		if err == nil {
			var profile entities.Profile
			if task.TaskRetailer != enums.Amazon {
				profile, err = queries.GetProfile(task.TaskProfileID)
			}
			if err == nil {
				validCardType := true
				if task.TaskRetailer != enums.Amazon {
					validCardType = common.ValidCardType([]byte(profile.CreditCard.CardNumber), task.TaskRetailer)
				}
				if validCardType {
					// Add task to store (if it already exists, this will return true)
					err = taskStore.AddTaskToStore(&task)
					if err == nil {
						taskPtr := taskStore.GetTask(task.TaskRetailer, taskID)
						if taskPtr != nil {
							task = *taskPtr
						}
						// Setting the stop flag to false before running the task
						taskStore.SetStopFlag(task.TaskRetailer, taskID, false)

						// If the Task is already running, then we're all set already
						if strings.Contains(task.TaskStatus, strings.ReplaceAll(enums.TaskIdle, " %s", "")) ||
							strings.Contains(task.TaskStatus, strings.ReplaceAll(enums.CheckingOutFailure, " %s", "")) ||
							strings.Contains(task.TaskStatus, strings.ReplaceAll(enums.CardDeclined, " %s", "")) ||
							strings.Contains(task.TaskStatus, strings.ReplaceAll(enums.CheckingOutSuccess, " %s", "")) ||
							strings.Contains(task.TaskStatus, strings.ReplaceAll(enums.TaskFailed, " %s", "")) {
							// Otherwise, start the Task
							taskStore.RunTask(task.TaskRetailer, task.ID)
						}
					} else {
						warnings = append(warnings, err.Error())
					}
				} else {
					warnings = append(warnings, errors.StartTaskInvalidCardError+task.TaskRetailer)
				}
			} else {
				warnings = append(warnings, err.Error())
			}
		} else {
			warnings = append(warnings, err.Error())
		}
	}

	if len(taskGroup.TaskIDs) == len(warnings) {
		err = e.New(errors.StartMonitorInvalidCardError + taskGroup.MonitorRetailer)
		monitorStore.StopMonitor(taskGroup)
	}

	return warnings, err
}

// StopTaskGroup sets the stop field for the given TaskGroup's Monitor and each Task in the group and returns true if successful
func (taskStore *TaskStore) StopTaskGroup(taskGroup *entities.TaskGroup) error {
	// Stop the task's TaskGroup
	_, err := monitorStore.StopMonitor(taskGroup)
	if err != nil {
		return err
	}

	// Set the tasks StopFlags to true
	for _, taskID := range taskGroup.TaskIDs {
		taskStore.SetStopFlag(taskGroup.MonitorRetailer, taskID, true)
	}

	return nil
}

func (taskStore *TaskStore) UpdateTask(newTask *entities.Task) error {
	task := taskStore.GetTask(newTask.TaskRetailer, newTask.ID)

	if task == nil {
		return e.New("task not found")
	}

	return taskStore.AddTaskToStore(newTask)
}

// StartTask runs the RunTask() function for the given Task and returns true if successful
func (taskStore *TaskStore) StartTask(task *entities.Task) error {
	if task.TaskRetailer != enums.Amazon {
		profile, err := queries.GetProfile(task.TaskProfileID)
		if err != nil {
			return err
		}

		if !common.ValidCardType([]byte(profile.CreditCard.CardNumber), task.TaskRetailer) {
			return e.New(errors.StartTaskInvalidCardError + task.TaskRetailer)
		}
	}

	taskGroup, err := queries.GetTaskGroup(task.TaskGroupID)
	if err != nil {
		return err
	}
	// Start the task's TaskGroup (if it's already running, this will return true)
	err = monitorStore.StartMonitor(&taskGroup)
	if err != nil {
		return err
	}

	// Add task to store (if it already exists, this will return true)
	err = taskStore.AddTaskToStore(task)
	if err != nil {
		return err
	}

	task = taskStore.GetTask(task.TaskRetailer, task.ID)

	// If the Task is already running, then we're all set already
	if !strings.Contains(task.TaskStatus, strings.ReplaceAll(enums.TaskIdle, " %s", "")) &&
		!strings.Contains(task.TaskStatus, strings.ReplaceAll(enums.CheckingOutFailure, " %s", "")) &&
		!strings.Contains(task.TaskStatus, strings.ReplaceAll(enums.CardDeclined, " %s", "")) &&
		!strings.Contains(task.TaskStatus, strings.ReplaceAll(enums.CheckingOutSuccess, " %s", "")) &&
		!strings.Contains(task.TaskStatus, strings.ReplaceAll(enums.TaskFailed, " %s", "")) {
		return nil
	}

	// Set the task's StopFlag to false before running the task
	taskStore.SetStopFlag(task.TaskRetailer, task.ID, false)
	taskStore.SetDontPublishEvents(task.TaskRetailer, task.ID, false)

	// Otherwise, start the Task
	taskStore.RunTask(task.TaskRetailer, task.ID)
	return nil
}

// StopTask sets the stop field for the given Task and returns true if successful
func (taskStore *TaskStore) StopTask(task *entities.Task) (bool, error) {
	if !taskStore.TasksRunning([]string{task.ID}, task.TaskRetailer) {
		return false, nil
	}
	return true, taskStore.SetStopFlag(task.TaskRetailer, task.ID, true)
}

// TasksRunning checks to see if any tasks in the taskGroup are running, if so it returns true
func (taskStore *TaskStore) TasksRunning(taskIDs []string, retailer enums.Retailer) bool {
	for _, taskID := range taskIDs {
		switch retailer {
		// Future sitescripts will have a case here
		case enums.Amazon:
			if amazonTask, ok := taskStore.AmazonTasks[taskID]; ok {
				if !amazonTask.Task.StopFlag {
					return true
				}
			}

		case enums.BestBuy:
			if bestbuyTask, ok := taskStore.BestbuyTasks[taskID]; ok {
				if !bestbuyTask.Task.StopFlag {
					return true
				}
			}
		case enums.BoxLunch:
			if boxlunchTask, ok := taskStore.BoxlunchTasks[taskID]; ok {
				if !boxlunchTask.Task.StopFlag {
					return true
				}
			}

		case enums.GameStop:
			if gamestopTask, ok := taskStore.GamestopTasks[taskID]; ok {
				if !gamestopTask.Task.StopFlag {
					return true
				}
			}

		case enums.Disney:
			if disneyTask, ok := taskStore.DisneyTasks[taskID]; ok {
				if !disneyTask.Task.StopFlag {
					return true
				}
			}

		case enums.HotTopic:
			if hottopicTask, ok := taskStore.HottopicTasks[taskID]; ok {
				if !hottopicTask.Task.StopFlag {
					return true
				}
			}

		case enums.Newegg:
			if neweggTask, ok := taskStore.NeweggTasks[taskID]; ok {
				if !neweggTask.Task.StopFlag {
					return true
				}
			}

		case enums.PokemonCenter:
			if pokemonCenterTask, ok := taskStore.PokemonCenterTasks[taskID]; ok {
				if !pokemonCenterTask.Task.StopFlag {
					return true
				}
			}

		case enums.Shopify:
			if shopifyTask, ok := taskStore.ShopifyTasks[taskID]; ok {
				if !shopifyTask.Task.StopFlag {
					return true
				}
			}

		case enums.Target:
			if targetTask, ok := taskStore.TargetTasks[taskID]; ok {
				if !targetTask.Task.StopFlag {
					return true
				}
			}

		case enums.Topps:
			if toppsTask, ok := taskStore.ToppsTasks[taskID]; ok {
				if !toppsTask.Task.StopFlag {
					return true
				}
			}

		case enums.Walmart:
			if walmartTask, ok := taskStore.WalmartTasks[taskID]; ok {
				if !walmartTask.Task.StopFlag {
					return true
				}
			}

		}
	}

	return false
}

func (taskStore *TaskStore) UpdateTaskProxy(task *entities.Task, proxy *entities.Proxy) bool {
	switch task.TaskRetailer {
	case enums.Amazon:
		if amazonTask, ok := taskStore.AmazonTasks[task.ID]; ok {
			amazonTask.Task.Proxy = proxy
		}
		return true

	case enums.BestBuy:
		if bestbuyTask, ok := taskStore.BestbuyTasks[task.ID]; ok {
			bestbuyTask.Task.Proxy = proxy
		}
		return true

	case enums.BoxLunch:
		if boxlunchTask, ok := taskStore.BoxlunchTasks[task.ID]; ok {
			boxlunchTask.Task.Proxy = proxy
		}
		return true

	case enums.Disney:
		if disneyTask, ok := taskStore.DisneyTasks[task.ID]; ok {
			disneyTask.Task.Proxy = proxy
		}
		return true

	case enums.GameStop:
		if gamestopTask, ok := taskStore.GamestopTasks[task.ID]; ok {
			gamestopTask.Task.Proxy = proxy
		}
		return true

	case enums.HotTopic:
		if hottopicTask, ok := taskStore.HottopicTasks[task.ID]; ok {
			hottopicTask.Task.Proxy = proxy
		}
		return true

	case enums.Newegg:
		if neweggTask, ok := taskStore.NeweggTasks[task.ID]; ok {
			neweggTask.Task.Proxy = proxy
		}
		return true

	case enums.PokemonCenter:
		if pokemonCenterTask, ok := taskStore.PokemonCenterTasks[task.ID]; ok {
			pokemonCenterTask.Task.Proxy = proxy
		}
		return true

	case enums.Shopify:
		if shopifyTask, ok := taskStore.ShopifyTasks[task.ID]; ok {
			shopifyTask.Task.Proxy = proxy
		}
		return true

	case enums.Target:
		if targetTask, ok := taskStore.TargetTasks[task.ID]; ok {
			targetTask.Task.Proxy = proxy
		}
		return true

	case enums.Topps:
		if toppsTask, ok := taskStore.ToppsTasks[task.ID]; ok {
			toppsTask.Task.Proxy = proxy
		}
		return true

	case enums.Walmart:
		if walmartTask, ok := taskStore.WalmartTasks[task.ID]; ok {
			walmartTask.Task.Proxy = proxy
		}
		return true

	}

	return false
}

func (taskStore *TaskStore) SetWalmartCardDetails(taskID string, cardInfo walmart.CardInfo) {
	if walmartTask, ok := taskStore.WalmartTasks[taskID]; ok {
		walmartTask.CardInfo = cardInfo
	}
}

func (taskStore *TaskStore) RunTask(retailer enums.Retailer, taskID string) {
	switch retailer {
	// Future sitescripts will have a case here
	case enums.Amazon:
		go taskStore.AmazonTasks[taskID].RunTask()

	case enums.BestBuy:
		go taskStore.BestbuyTasks[taskID].RunTask()

	case enums.BoxLunch:
		go taskStore.BoxlunchTasks[taskID].RunTask()

	case enums.Disney:
		go taskStore.DisneyTasks[taskID].RunTask()

	case enums.GameStop:
		go taskStore.GamestopTasks[taskID].RunTask()

	case enums.HotTopic:
		go taskStore.HottopicTasks[taskID].RunTask()

	case enums.Newegg:
		go taskStore.NeweggTasks[taskID].RunTask()

	case enums.PokemonCenter:
		go taskStore.PokemonCenterTasks[taskID].RunTask()

	case enums.Shopify:
		go taskStore.ShopifyTasks[taskID].RunTask()

	case enums.Target:
		go taskStore.TargetTasks[taskID].RunTask()

	case enums.Topps:
		go taskStore.ToppsTasks[taskID].RunTask()

	case enums.Walmart:
		go taskStore.WalmartTasks[taskID].RunTask()

	}
}

var taskStore *TaskStore

// InitTaskStore initializes the singleton instance of the TaskStore
func InitTaskStore(eventBus *events.EventBus) {
	taskStore = &TaskStore{
		AmazonTasks:        make(map[string]*amazon.Task),
		BestbuyTasks:       make(map[string]*bestbuy.Task),
		BoxlunchTasks:      make(map[string]*boxlunch.Task),
		DisneyTasks:        make(map[string]*disney.Task),
		GamestopTasks:      make(map[string]*gamestop.Task),
		HottopicTasks:      make(map[string]*hottopic.Task),
		NeweggTasks:        make(map[string]*newegg.Task),
		PokemonCenterTasks: make(map[string]*pokemoncenter.Task),
		ShopifyTasks:       make(map[string]*shopify.Task),
		TargetTasks:        make(map[string]*target.Task),
		ToppsTasks:         make(map[string]*topps.Task),
		WalmartTasks:       make(map[string]*walmart.Task),

		EventBus: eventBus,
	}
}

// GetTaskStatuses returns a list of tasks with the most up to date status
func GetTaskStatuses() map[string]string {
	taskStatuses := make(map[string]string)

	for taskID, task := range taskStore.AmazonTasks {
		taskStatuses[taskID] = task.Task.Task.TaskStatus
	}
	for taskID, task := range taskStore.BestbuyTasks {
		taskStatuses[taskID] = task.Task.Task.TaskStatus
	}
	for taskID, task := range taskStore.DisneyTasks {
		taskStatuses[taskID] = task.Task.Task.TaskStatus
	}
	for taskID, task := range taskStore.GamestopTasks {
		taskStatuses[taskID] = task.Task.Task.TaskStatus
	}
	for taskID, task := range taskStore.HottopicTasks {
		taskStatuses[taskID] = task.Task.Task.TaskStatus
	}
	for taskID, task := range taskStore.NeweggTasks {
		taskStatuses[taskID] = task.Task.Task.TaskStatus
	}
	for taskID, task := range taskStore.PokemonCenterTasks {
		taskStatuses[taskID] = task.Task.Task.TaskStatus
	}
	for taskID, task := range taskStore.ShopifyTasks {
		taskStatuses[taskID] = task.Task.Task.TaskStatus
	}
	for taskID, task := range taskStore.TargetTasks {
		taskStatuses[taskID] = task.Task.Task.TaskStatus
	}
	for taskID, task := range taskStore.ToppsTasks {
		taskStatuses[taskID] = task.Task.Task.TaskStatus
	}
	for taskID, task := range taskStore.WalmartTasks {
		taskStatuses[taskID] = task.Task.Task.TaskStatus
	}

	return taskStatuses
}

// GetTaskStore returns the singleton instance of the EventBus
func GetTaskStore() *TaskStore {
	return taskStore
}
