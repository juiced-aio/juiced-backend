package stores

func InitStores() error {
	var err error

	err = taskStore.Init()
	if err != nil {
		return err
	}

	err = taskGroupStore.Init()
	if err != nil {
		return err
	}

	err = profileStore.Init()
	if err != nil {
		return err
	}

	err = profileGroupStore.Init()
	if err != nil {
		return err
	}

	err = proxyGroupStore.Init()
	if err != nil {
		return err
	}

	return nil
}
