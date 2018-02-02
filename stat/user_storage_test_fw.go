package stat

import (
	"errors"
	"fmt"
	// "github.com/KyberNetwork/reserve-data/common"
	// ethereum "github.com/ethereum/go-ethereum/common"
)

// This test type enforces necessary logic required for a stat storage.
// - It requires an actual storage instance to be able to run the tests.
// - It DOESNT do any tear up or tear down processes.
// - Each of its functions is for one test and will return non-nil error
// if the test didn't pass.
// - It is supposed to be used in a package that has the knowledge
// of actual storage being used as this interface.
// Eg. It should be used in cmd package where we decide to use
// bolt (for example) as the storage for stat storage
type UserStorageTest struct {
	storage UserStorage
}

func NewUserStorageTest(storage UserStorage) *UserStorageTest {
	return &UserStorageTest{storage}
}

func (self *UserStorageTest) TestUpdateAddressCategory() error {
	lowercaseAddr := "0x8180a5ca4e3b94045e05a9313777955f7518d757"
	lowercaseCat := "0x4a"
	addr := "0x8180a5CA4E3B94045e05A9313777955f7518D757"
	cat := "0x4A"
	if err := self.storage.UpdateAddressCategory(addr, cat); err != nil {
		return err
	}
	gotCat, err := self.storage.GetCategory(addr)
	if err != nil {
		return err
	}
	if gotCat != lowercaseCat {
		return errors.New(fmt.Sprintf("Got unexpected category. Expected(%s) Got(%s)",
			lowercaseCat, gotCat))
	}
	gotCat, err = self.storage.GetCategory(lowercaseAddr)
	if err != nil {
		return err
	}
	if gotCat != lowercaseCat {
		return errors.New(fmt.Sprintf("Got unexpected category. Expected(%s) Got(%s)",
			lowercaseCat, gotCat))
	}
	user, _, err := self.storage.GetUserOfAddress(lowercaseAddr)
	// initialy user is identical to the address
	if err != nil {
		return err
	}
	if user != lowercaseAddr {
		return errors.New(fmt.Sprintf("Got unexpected user. Expected(%s) Got(%s)",
			user, lowercaseAddr))
	}
	addresses, _, err := self.storage.GetAddressesOfUser(user)
	if err != nil {
		return err
	}
	if addresses[0] != lowercaseAddr {
		return errors.New(fmt.Sprintf("Got unexpected addresses. Expected(%v) Got(%v)",
			addresses, []string{lowercaseAddr}))
	}
	return nil
}

func (self *UserStorageTest) TestUpdateUserAddressesThenUpdateAddressCategory() error {
	email := "victor@kyber.network"
	addr1 := "0x8180a5ca4e3b94045e05a9313777955f7518d757"
	time1 := uint64(1520825136556)
	addr2 := "0xcbac9e86e0b7160f1a8e4835ad01dd51c514afce"
	time2 := uint64(1520825136557)
	addr3 := "0x0ccd5bd8eb6822d357d7aef833274502e8b4b8ac"
	time3 := uint64(1520825136558)
	cat := "0x0000000000000000000000000000000000000000000000000000000000000004"

	self.storage.UpdateUserAddresses(
		email, []string{addr1, addr3}, []uint64{time1, time3},
	)
	// test if pending addresses are correct
	pendingAddrs, err := self.storage.GetPendingAddresses()
	if err != nil {
		return err
	}
	expectedAddresses := map[string]uint64{
		addr1: time1,
		addr3: time3,
	}
	if len(pendingAddrs) != len(expectedAddresses) {
		return errors.New(
			fmt.Sprintf("Expected to get %d addresses, got %d addresses", len(expectedAddresses), len(pendingAddrs)))
	}
	for _, addr := range pendingAddrs {
		if _, found := expectedAddresses[addr]; !found {
			return errors.New(fmt.Sprintf("Expected to find %s, got not found", addr))
		}
	}
	self.storage.UpdateUserAddresses(
		email, []string{addr1, addr2}, []uint64{time1, time2},
	)
	// test if pending addresses are correct
	pendingAddrs, err = self.storage.GetPendingAddresses()
	if err != nil {
		return err
	}
	expectedAddresses = map[string]uint64{
		addr1: time1,
		addr2: time2,
	}
	if len(pendingAddrs) != len(expectedAddresses) {
		return errors.New(
			fmt.Sprintf("Expected to get %d addresses, got %d addresses", len(expectedAddresses), len(pendingAddrs)))
	}
	for _, addr := range pendingAddrs {
		if _, found := expectedAddresses[addr]; !found {
			return errors.New(fmt.Sprintf("Expected to find %s, got not found", addr))
		}
	}
	// Start receiving cat logs
	self.storage.UpdateAddressCategory(addr1, cat)
	self.storage.UpdateUserAddresses(
		email, []string{addr1, addr2}, []uint64{time1, time2},
	)
	// test if pending addresses are correct
	pendingAddrs, err = self.storage.GetPendingAddresses()
	if err != nil {
		return err
	}
	expectedAddresses = map[string]uint64{
		addr2: time2,
	}
	if len(pendingAddrs) != len(expectedAddresses) {
		return errors.New(
			fmt.Sprintf("Expected to get %d addresses, got %d addresses", len(expectedAddresses), len(pendingAddrs)))
	}
	for _, addr := range pendingAddrs {
		if _, found := expectedAddresses[addr]; !found {
			return errors.New(fmt.Sprintf("Expected to find %s, got not found", addr))
		}
	}
	self.storage.UpdateAddressCategory(addr2, cat)

	gotAddresses, gotTimes, err := self.storage.GetAddressesOfUser(email)
	if err != nil {
		return err
	}
	// test addresses of user
	expectedAddresses = map[string]uint64{
		addr1: time1,
		addr2: time2,
	}
	if len(gotAddresses) != len(expectedAddresses) {
		return errors.New(
			fmt.Sprintf("Expected to get %d addresses, got %d addresses", len(expectedAddresses), len(gotAddresses)))
	}
	for i, addr := range gotAddresses {
		if _, found := expectedAddresses[addr]; !found {
			return errors.New(fmt.Sprintf("Expected to find %s, got not found", addr))
		}
		if expectedAddresses[addr] != gotTimes[i] {
			return errors.New(fmt.Sprintf("Expected timestamp %d, got %d", expectedAddresses[addr], gotTimes[i]))
		}
	}
	gotUser, gotTime, err := self.storage.GetUserOfAddress(addr1)
	if err != nil {
		return err
	}
	if gotUser != email {
		return errors.New(fmt.Sprintf("Expected to get %s, got %s", email, gotUser))
	}
	if gotTime != time1 {
		return errors.New(fmt.Sprintf("Expected to get %d, got %d", time1, gotTime))
	}
	gotUser, gotTime, err = self.storage.GetUserOfAddress(addr2)
	if err != nil {
		return err
	}
	if gotUser != email {
		return errors.New(fmt.Sprintf("Expected to get %s, got %s", email, gotUser))
	}
	if gotTime != time2 {
		return errors.New(fmt.Sprintf("Expected to get %d, got %d", time2, gotTime))
	}
	return nil
}

func (self *UserStorageTest) TestUpdateAddressCategoryThenUpdateUserAddresses() error {
	email := "Victor@kyber.network"
	lowercaseEmail := "victor@kyber.network"
	addr1 := "0x8180a5CA4E3B94045e05A9313777955f7518D757"
	time1 := uint64(1520825136556)
	lowercaseAddr1 := "0x8180a5ca4e3b94045e05a9313777955f7518d757"
	addr2 := "0xcbac9e86e0b7160f1a8e4835ad01dd51c514afce"
	time2 := uint64(1520825136557)
	cat := "0x4A"

	self.storage.UpdateAddressCategory(addr1, cat)
	self.storage.UpdateAddressCategory(addr2, cat)
	err := self.storage.UpdateUserAddresses(
		email, []string{addr1, addr2}, []uint64{time1, time2},
	)
	if err != nil {
		return err
	}
	gotAddresses, gotTimes, err := self.storage.GetAddressesOfUser(lowercaseEmail)
	if err != nil {
		return err
	}
	expectedAddresses := map[string]uint64{
		lowercaseAddr1: time1,
		addr2:          time2,
	}
	if len(gotAddresses) != len(expectedAddresses) {
		return errors.New(
			fmt.Sprintf("Expected to get %d addresses, got %d addresses", len(expectedAddresses), len(gotAddresses)))
	}
	for i, addr := range gotAddresses {
		if _, found := expectedAddresses[addr]; !found {
			return errors.New(fmt.Sprintf("Expected to find %s, got not found", addr))
		}
		if gotTimes[i] != expectedAddresses[addr] {
			return errors.New(fmt.Sprintf("Expected %d, found %d", expectedAddresses[addr], gotTimes[i]))
		}
	}
	gotUser, gotTime, err := self.storage.GetUserOfAddress(addr1)
	if err != nil {
		return err
	}
	if gotUser != lowercaseEmail {
		return errors.New(fmt.Sprintf("Expected to get %s, got %s", lowercaseEmail, gotUser))
	}
	if gotTime != time1 {
		return errors.New(fmt.Sprintf("Expected to get %d, got %d", time1, gotTime))
	}
	gotUser, gotTime, err = self.storage.GetUserOfAddress(addr2)
	if err != nil {
		return err
	}
	if gotUser != lowercaseEmail {
		return errors.New(fmt.Sprintf("Expected to get %s, got %s", lowercaseEmail, gotUser))
	}
	if gotTime != time2 {
		return errors.New(fmt.Sprintf("Expected to get %d, got %d", time2, gotTime))
	}
	return nil
}