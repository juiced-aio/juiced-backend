package stores

func InitStores() error {
	var err error

	err = InitAccountStore()
	if err != nil {
		return err
	}

	err = InitProfileStore()
	if err != nil {
		return err
	}

	err = InitProfileGroupStore()
	if err != nil {
		return err
	}

	err = InitProxyGroupStore()
	if err != nil {
		return err
	}

	err = InitTaskStore()
	if err != nil {
		return err
	}

	err = InitTaskGroupStore()
	if err != nil {
		return err
	}

	return nil
}
