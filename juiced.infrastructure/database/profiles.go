package database

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	"backend.juicedbot.io/juiced.infrastructure/util"
)

// GetAllProfiles returns all Profile objects from the database
func GetAllProfiles() ([]entities.Profile, error) {
	profiles := []entities.Profile{}
	if database == nil {
		return profiles, &DatabaseNotInitializedError{}
	}

	rows, err := database.Queryx("SELECT * FROM profiles")
	if err != nil {
		return profiles, err
	}

	defer rows.Close()
	for rows.Next() {
		var encryptedEmail string
		var encryptedPhoneNumber string

		tempProfile := entities.Profile{}
		err = rows.StructScan(&tempProfile)
		if err != nil {
			return profiles, err
		}

		decryptedEmail, err := util.Aes256Decrypt(tempProfile.Email, enums.UserKey)
		if err == nil {
			tempProfile.Email = decryptedEmail
		} else {
			encryptedEmail, err = util.Aes256Encrypt(tempProfile.Email, enums.UserKey)
			if err != nil {
				return profiles, err
			}
		}
		decryptedPhoneNumber, err := util.Aes256Decrypt(tempProfile.PhoneNumber, enums.UserKey)
		if err == nil {
			tempProfile.PhoneNumber = decryptedPhoneNumber
		} else {
			encryptedPhoneNumber, err = util.Aes256Encrypt(tempProfile.PhoneNumber, enums.UserKey)
			if err != nil {
				return profiles, err
			}
		}

		if encryptedEmail != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE profiles SET email = "%v" WHERE ID = "%v"`, encryptedEmail, tempProfile.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedPhoneNumber != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE profiles SET phoneNumber = "%v" WHERE ID = "%v"`, encryptedPhoneNumber, tempProfile.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		tempProfile, err = GetProfileInfo(tempProfile)
		if err != nil {
			return profiles, err
		}
		if tempProfile.ID != "" && tempProfile.Name != "" {
			profiles = append(profiles, tempProfile)
		}
	}

	sort.SliceStable(profiles, func(i, j int) bool {
		return profiles[i].CreationDate < profiles[j].CreationDate
	})

	return profiles, err
}

// GetProfile returns the Profile object from the database with the given ID (if it exists)
func GetProfile(ID string) (entities.Profile, error) {
	profile := entities.Profile{}
	if database == nil {
		return profile, &DatabaseNotInitializedError{}
	}

	statement, err := database.Preparex("SELECT * FROM profiles WHERE ID = @p1")
	if err != nil {
		return profile, err
	}

	rows, err := statement.Queryx(ID)
	if err != nil {
		return profile, err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&profile)
		if err != nil {
			return profile, err
		}
	}

	return GetProfileInfo(profile)
}

// GetProfileByName returns the Profile object from the database with the given name (if it exists)
func GetProfileByName(name string) (entities.Profile, error) {
	profile := entities.Profile{}
	if database == nil {
		return profile, &DatabaseNotInitializedError{}
	}

	statement, err := database.Preparex("SELECT * FROM profiles WHERE Name = @p1")
	if err != nil {
		return profile, err
	}

	rows, err := statement.Queryx(name)
	if err != nil {
		return profile, err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&profile)
		if err != nil {
			return profile, err
		}
	}

	return GetProfileInfo(profile)
}

func CreateProfile(profile entities.Profile) error {
	if database == nil {
		return &DatabaseNotInitializedError{}
	}

	encryptedEmail, err := util.Aes256Encrypt(profile.Email, enums.UserKey)
	if err != nil {
		return err
	}
	encryptedPhoneNumber, err := util.Aes256Encrypt(profile.PhoneNumber, enums.UserKey)
	if err != nil {
		return err
	}

	statement, err := database.Preparex(`INSERT INTO profiles (ID, profileGroupIDsJoined, name, email, phoneNumber, creationDate) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	_, err = statement.Exec(profile.ID, profile.ProfileGroupIDsJoined, profile.Name, encryptedEmail, encryptedPhoneNumber, profile.CreationDate)
	if err != nil {
		return err
	}

	return CreateProfileInfos(profile)
}

func RemoveProfile(ID string) error {
	if database == nil {
		return &DatabaseNotInitializedError{}
	}

	statement, err := database.Preparex(`DELETE FROM profiles WHERE ID = @p1`)
	if err != nil {
		return err
	}
	_, err = statement.Exec(ID)
	if err != nil {
		return err
	}

	return DeleteProfileInfos(ID)
}

func UpdateProfile(ID string, newProfile entities.Profile) error {
	err := RemoveProfile(ID)
	if err != nil {
		return err
	}
	newProfile.ProfileGroupIDsJoined = strings.Join(newProfile.ProfileGroupIDs, ",")
	return CreateProfile(newProfile)
}

func CreateShippingAddresses(profile entities.Profile) error {
	if database == nil {
		return &DatabaseNotInitializedError{}
	}

	encryptedValues, err := util.EncryptValues(enums.UserKey, profile.ShippingAddress.FirstName, profile.ShippingAddress.LastName, profile.ShippingAddress.Address1, profile.ShippingAddress.Address2, profile.ShippingAddress.City, profile.ShippingAddress.ZipCode, profile.ShippingAddress.StateCode, profile.ShippingAddress.CountryCode)
	if err != nil {
		return err
	}

	statement, err := database.Preparex(`INSERT INTO shippingAddresses (ID, profileID, firstName, lastName, address1, address2, city, zipCode, stateCode, countryCode) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	_, err = statement.Exec(profile.ShippingAddress.ID, profile.ID, encryptedValues[0], encryptedValues[1], encryptedValues[2], encryptedValues[3], encryptedValues[4], encryptedValues[5], encryptedValues[6], encryptedValues[7])
	if err != nil {
		return err
	}

	return err
}
func CreateBillingAddresses(profile entities.Profile) error {
	if database == nil {
		return &DatabaseNotInitializedError{}
	}

	encryptedValues, err := util.EncryptValues(enums.UserKey, profile.BillingAddress.FirstName, profile.BillingAddress.LastName, profile.BillingAddress.Address1, profile.BillingAddress.Address2, profile.BillingAddress.City, profile.BillingAddress.ZipCode, profile.BillingAddress.StateCode, profile.BillingAddress.CountryCode)
	if err != nil {
		return err
	}

	statement, err := database.Preparex(`INSERT INTO billingAddresses (ID, profileID, firstName, lastName, address1, address2, city, zipCode, stateCode, countryCode) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	_, err = statement.Exec(profile.BillingAddress.ID, profile.ID, encryptedValues[0], encryptedValues[1], encryptedValues[2], encryptedValues[3], encryptedValues[4], encryptedValues[5], encryptedValues[6], encryptedValues[7])
	if err != nil {
		return err
	}

	return err
}
func CreateCards(profile entities.Profile) error {
	if database == nil {
		return &DatabaseNotInitializedError{}
	}

	encryptedValues, err := util.EncryptValues(enums.UserKey, profile.CreditCard.CardholderName, profile.CreditCard.CardNumber, profile.CreditCard.ExpMonth, profile.CreditCard.ExpYear, profile.CreditCard.CVV, profile.CreditCard.CardType)
	if err != nil {
		return err
	}

	statement, err := database.Preparex(`INSERT INTO cards (ID, profileID, cardHolderName, cardNumber, expMonth, expYear, cvv, cardType) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	_, err = statement.Exec(profile.CreditCard.ID, profile.ID, encryptedValues[0], encryptedValues[1], encryptedValues[2], encryptedValues[3], encryptedValues[4], encryptedValues[5])
	if err != nil {
		return err
	}

	return err
}

func CreateProfileInfos(profile entities.Profile) error {
	err := CreateShippingAddresses(profile)
	if err != nil {
		return err
	}
	err = CreateBillingAddresses(profile)
	if err != nil {
		return err
	}
	return CreateCards(profile)
}

func DeleteShippingAddresses(ID string) error {
	if database == nil {
		return &DatabaseNotInitializedError{}
	}

	statement, err := database.Preparex(`DELETE FROM shippingAddresses WHERE profileID = @p1`)
	if err != nil {
		return err
	}
	_, err = statement.Exec(ID)

	return err
}

func DeleteBillingAddresses(ID string) error {
	if database == nil {
		return &DatabaseNotInitializedError{}
	}

	statement, err := database.Preparex(`DELETE FROM billingAddresses WHERE profileID = @p1`)
	if err != nil {
		return err
	}
	_, err = statement.Exec(ID)

	return err
}

func DeleteCards(ID string) error {
	if database == nil {
		return &DatabaseNotInitializedError{}
	}

	statement, err := database.Preparex(`DELETE FROM cards WHERE profileID = @p1`)
	if err != nil {
		return err
	}
	_, err = statement.Exec(ID)

	return err
}

func DeleteProfileInfos(ID string) error {
	err := DeleteShippingAddresses(ID)
	if err != nil {
		return err
	}
	err = DeleteBillingAddresses(ID)
	if err != nil {
		return err
	}
	return DeleteCards(ID)
}

func GetShippingAddress(profile entities.Profile) (entities.Profile, error) {
	if database == nil {
		return profile, &DatabaseNotInitializedError{}
	}

	statement, err := database.Preparex(`SELECT * FROM shippingAddresses WHERE profileID = @p1`)
	if err != nil {
		return profile, err
	}
	rows, err := statement.Queryx(profile.ID)
	if err != nil {
		return profile, err
	}

	defer rows.Close()
	for rows.Next() {
		var encryptedFirstName string
		var encryptedLastName string
		var encryptedAddress1 string
		var encryptedAddress2 string
		var encryptedCity string
		var encryptedZipCode string
		var encryptedStateCode string
		var encryptedCountryCode string

		err = rows.StructScan(&profile.ShippingAddress)
		if err != nil {
			return profile, err
		}

		decryptedFirstName, err := util.Aes256Decrypt(profile.ShippingAddress.FirstName, enums.UserKey)
		if err == nil {
			profile.ShippingAddress.FirstName = decryptedFirstName
		} else {
			encryptedFirstName, err = util.Aes256Encrypt(profile.ShippingAddress.FirstName, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedLastName, err := util.Aes256Decrypt(profile.ShippingAddress.LastName, enums.UserKey)
		if err == nil {
			profile.ShippingAddress.LastName = decryptedLastName
		} else {
			encryptedLastName, err = util.Aes256Encrypt(profile.ShippingAddress.LastName, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedAddress1, err := util.Aes256Decrypt(profile.ShippingAddress.Address1, enums.UserKey)
		if err == nil {
			profile.ShippingAddress.Address1 = decryptedAddress1
		} else {
			encryptedAddress1, err = util.Aes256Encrypt(profile.ShippingAddress.Address1, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedAddress2, err := util.Aes256Decrypt(profile.ShippingAddress.Address2, enums.UserKey)
		if err == nil {
			profile.ShippingAddress.Address2 = decryptedAddress2
		} else {
			encryptedAddress2, err = util.Aes256Encrypt(profile.ShippingAddress.Address2, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedCity, err := util.Aes256Decrypt(profile.ShippingAddress.City, enums.UserKey)
		if err == nil {
			profile.ShippingAddress.City = decryptedCity
		} else {
			encryptedCity, err = util.Aes256Encrypt(profile.ShippingAddress.City, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedZipCode, err := util.Aes256Decrypt(profile.ShippingAddress.ZipCode, enums.UserKey)
		if err == nil {
			profile.ShippingAddress.ZipCode = decryptedZipCode
		} else {
			encryptedZipCode, err = util.Aes256Encrypt(profile.ShippingAddress.ZipCode, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedStateCode, err := util.Aes256Decrypt(profile.ShippingAddress.StateCode, enums.UserKey)
		if err == nil {
			profile.ShippingAddress.StateCode = decryptedStateCode
		} else {
			encryptedStateCode, err = util.Aes256Encrypt(profile.ShippingAddress.StateCode, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedCountryCode, err := util.Aes256Decrypt(profile.ShippingAddress.CountryCode, enums.UserKey)
		if err == nil {
			profile.ShippingAddress.CountryCode = decryptedCountryCode
		} else {
			encryptedCountryCode, err = util.Aes256Encrypt(profile.ShippingAddress.CountryCode, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}

		if encryptedFirstName != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE shippingAddresses SET firstName = "%v" WHERE ID = "%v"`, encryptedFirstName, profile.ShippingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedLastName != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE shippingAddresses SET lastName = "%v" WHERE ID = "%v"`, encryptedLastName, profile.ShippingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedAddress1 != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE shippingAddresses SET address1 = "%v" WHERE ID = "%v"`, encryptedAddress1, profile.ShippingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedAddress2 != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE shippingAddresses SET address2 = "%v" WHERE ID = "%v"`, encryptedAddress2, profile.ShippingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedCity != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE shippingAddresses SET city = "%v" WHERE ID = "%v"`, encryptedCity, profile.ShippingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}

			}()
		}

		if encryptedZipCode != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE shippingAddresses SET zipCode = "%v" WHERE ID = "%v"`, encryptedZipCode, profile.ShippingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedStateCode != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE shippingAddresses SET stateCode = "%v" WHERE ID = "%v"`, encryptedStateCode, profile.ShippingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedCountryCode != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE shippingAddresses SET countryCode = "%v" WHERE ID = "%v"`, encryptedCountryCode, profile.ShippingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

	}

	return profile, err
}

func GetBillingAddress(profile entities.Profile) (entities.Profile, error) {
	if database == nil {
		return profile, &DatabaseNotInitializedError{}
	}

	statement, err := database.Preparex(`SELECT * FROM billingAddresses WHERE profileID = @p1`)
	if err != nil {
		return profile, err
	}
	rows, err := statement.Queryx(profile.ID)
	if err != nil {
		return profile, err
	}

	defer rows.Close()
	for rows.Next() {
		var encryptedFirstName string
		var encryptedLastName string
		var encryptedAddress1 string
		var encryptedAddress2 string
		var encryptedCity string
		var encryptedZipCode string
		var encryptedStateCode string
		var encryptedCountryCode string

		err = rows.StructScan(&profile.BillingAddress)
		if err != nil {
			return profile, err
		}

		decryptedFirstName, err := util.Aes256Decrypt(profile.BillingAddress.FirstName, enums.UserKey)
		if err == nil {
			profile.BillingAddress.FirstName = decryptedFirstName
		} else {
			encryptedFirstName, err = util.Aes256Encrypt(profile.BillingAddress.FirstName, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedLastName, err := util.Aes256Decrypt(profile.BillingAddress.LastName, enums.UserKey)
		if err == nil {
			profile.BillingAddress.LastName = decryptedLastName
		} else {
			encryptedLastName, err = util.Aes256Encrypt(profile.BillingAddress.LastName, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedAddress1, err := util.Aes256Decrypt(profile.BillingAddress.Address1, enums.UserKey)
		if err == nil {
			profile.BillingAddress.Address1 = decryptedAddress1
		} else {
			encryptedAddress1, err = util.Aes256Encrypt(profile.BillingAddress.Address1, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedAddress2, err := util.Aes256Decrypt(profile.BillingAddress.Address2, enums.UserKey)
		if err == nil {
			profile.BillingAddress.Address2 = decryptedAddress2
		} else {
			encryptedAddress2, err = util.Aes256Encrypt(profile.BillingAddress.Address2, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedCity, err := util.Aes256Decrypt(profile.BillingAddress.City, enums.UserKey)
		if err == nil {
			profile.BillingAddress.City = decryptedCity
		} else {
			encryptedCity, err = util.Aes256Encrypt(profile.BillingAddress.City, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedZipCode, err := util.Aes256Decrypt(profile.BillingAddress.ZipCode, enums.UserKey)
		if err == nil {
			profile.BillingAddress.ZipCode = decryptedZipCode
		} else {
			encryptedZipCode, err = util.Aes256Encrypt(profile.BillingAddress.ZipCode, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedStateCode, err := util.Aes256Decrypt(profile.BillingAddress.StateCode, enums.UserKey)
		if err == nil {
			profile.BillingAddress.StateCode = decryptedStateCode
		} else {
			encryptedStateCode, err = util.Aes256Encrypt(profile.BillingAddress.StateCode, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedCountryCode, err := util.Aes256Decrypt(profile.BillingAddress.CountryCode, enums.UserKey)
		if err == nil {
			profile.BillingAddress.CountryCode = decryptedCountryCode
		} else {
			encryptedCountryCode, err = util.Aes256Encrypt(profile.BillingAddress.CountryCode, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}

		if encryptedFirstName != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE billingAddresses SET firstName = "%v" WHERE ID = "%v"`, encryptedFirstName, profile.BillingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedLastName != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE billingAddresses SET lastName = "%v" WHERE ID = "%v"`, encryptedLastName, profile.BillingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedAddress1 != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE billingAddresses SET address1 = "%v" WHERE ID = "%v"`, encryptedAddress1, profile.BillingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedAddress2 != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE billingAddresses SET address2 = "%v" WHERE ID = "%v"`, encryptedAddress2, profile.BillingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedCity != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE billingAddresses SET city = "%v" WHERE ID = "%v"`, encryptedCity, profile.BillingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}

			}()
		}

		if encryptedZipCode != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE billingAddresses SET zipCode = "%v" WHERE ID = "%v"`, encryptedZipCode, profile.BillingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedStateCode != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE billingAddresses SET stateCode = "%v" WHERE ID = "%v"`, encryptedStateCode, profile.BillingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedCountryCode != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE billingAddresses SET countryCode = "%v" WHERE ID = "%v"`, encryptedCountryCode, profile.BillingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

	}

	return profile, err
}

func GetCard(profile entities.Profile) (entities.Profile, error) {
	if database == nil {
		return profile, &DatabaseNotInitializedError{}
	}

	statement, err := database.Preparex(`SELECT * FROM cards WHERE profileID = @p1`)
	if err != nil {
		return profile, err
	}
	rows, err := statement.Queryx(profile.ID)
	if err != nil {
		return profile, err
	}

	defer rows.Close()
	for rows.Next() {
		var encryptedCardholderName string
		var encryptedCardNumber string
		var encryptedExpMonth string
		var encryptedExpYear string
		var encryptedCVV string
		var encryptedCardType string

		err = rows.StructScan(&profile.CreditCard)
		if err != nil {
			return profile, err
		}

		decryptedCardholderName, err := util.Aes256Decrypt(profile.CreditCard.CardholderName, enums.UserKey)
		if err == nil {
			profile.CreditCard.CardholderName = decryptedCardholderName
		} else {
			encryptedCardholderName, err = util.Aes256Encrypt(profile.CreditCard.CardholderName, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedCardNumber, err := util.Aes256Decrypt(profile.CreditCard.CardNumber, enums.UserKey)
		if err == nil {
			profile.CreditCard.CardNumber = decryptedCardNumber
		} else {
			encryptedCardNumber, err = util.Aes256Encrypt(profile.CreditCard.CardNumber, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedExpMonth, err := util.Aes256Decrypt(profile.CreditCard.ExpMonth, enums.UserKey)
		if err == nil {
			profile.CreditCard.ExpMonth = decryptedExpMonth
		} else {
			encryptedExpMonth, err = util.Aes256Encrypt(profile.CreditCard.ExpMonth, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedExpYear, err := util.Aes256Decrypt(profile.CreditCard.ExpYear, enums.UserKey)
		if err == nil {
			profile.CreditCard.ExpYear = decryptedExpYear
		} else {
			encryptedExpYear, err = util.Aes256Encrypt(profile.CreditCard.ExpYear, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedCVV, err := util.Aes256Decrypt(profile.CreditCard.CVV, enums.UserKey)
		if err == nil {
			profile.CreditCard.CVV = decryptedCVV
		} else {
			encryptedCVV, err = util.Aes256Encrypt(profile.CreditCard.CVV, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedCardType, err := util.Aes256Decrypt(profile.CreditCard.CardType, enums.UserKey)
		if err == nil {
			profile.CreditCard.CardType = decryptedCardType
		} else {
			encryptedCardType, err = util.Aes256Encrypt(profile.CreditCard.CardType, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}

		if encryptedCardholderName != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE cards SET cardHolderName = "%v" WHERE ID = "%v"`, encryptedCardholderName, profile.CreditCard.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedCardNumber != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE cards SET cardNumber = "%v" WHERE ID = "%v"`, encryptedCardNumber, profile.CreditCard.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedExpMonth != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE cards SET expMonth = "%v" WHERE ID = "%v"`, encryptedExpMonth, profile.CreditCard.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedExpYear != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE cards SET expYear = "%v" WHERE ID = "%v"`, encryptedExpYear, profile.CreditCard.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedCVV != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE cards SET cvv = "%v" WHERE ID = "%v"`, encryptedCVV, profile.CreditCard.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}

			}()
		}

		if encryptedCardType != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE cards SET cardType = "%v" WHERE ID = "%v"`, encryptedCardType, profile.CreditCard.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}
	}

	return profile, err
}

func GetProfileInfo(profile entities.Profile) (entities.Profile, error) {
	if database == nil {
		return profile, &DatabaseNotInitializedError{}
	}

	if profile.ProfileGroupIDsJoined != "" {
		profile.ProfileGroupIDs = strings.Split(profile.ProfileGroupIDsJoined, ",")
	}

	var (
		encryptedEmail       string
		encryptedPhoneNumber string
	)

	decryptedEmail, err := util.Aes256Decrypt(profile.Email, enums.UserKey)
	if err == nil {
		profile.Email = decryptedEmail
	} else {
		encryptedEmail, err = util.Aes256Encrypt(profile.Email, enums.UserKey)
		if err != nil {
			return profile, err
		}
	}

	decryptedPhoneNumber, err := util.Aes256Decrypt(profile.PhoneNumber, enums.UserKey)
	if err == nil {
		profile.PhoneNumber = decryptedPhoneNumber
	} else {
		encryptedPhoneNumber, err = util.Aes256Encrypt(profile.PhoneNumber, enums.UserKey)
		if err != nil {
			return profile, err
		}
	}

	if encryptedEmail != "" {
		go func() {
			for {
				_, err = database.Exec(fmt.Sprintf(`UPDATE profiles SET email = "%v" WHERE ID = "%v"`, encryptedEmail, profile.ID))
				if err != nil {
					time.Sleep(1 * time.Second)
					continue
				} else {
					break
				}
			}
		}()
	}
	if encryptedPhoneNumber != "" {
		go func() {
			for {
				_, err = database.Exec(fmt.Sprintf(`UPDATE profiles SET phoneNumber = "%v" WHERE ID = "%v"`, encryptedPhoneNumber, profile.ID))
				if err != nil {
					time.Sleep(1 * time.Second)
					continue
				} else {
					break
				}
			}
		}()
	}

	profile, err = GetShippingAddress(profile)
	if err != nil {
		return profile, err
	}
	profile, err = GetBillingAddress(profile)
	if err != nil {
		return profile, err
	}

	return GetCard(profile)
}
