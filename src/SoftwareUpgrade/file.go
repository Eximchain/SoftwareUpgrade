package softwareupgrade

import "io/ioutil"

// ReadDataFromFile reads the contents of the given filename into a byte array and returns it.
func ReadDataFromFile(filename string) ([]byte, error) {
	result, err := ioutil.ReadFile(filename)
	return result, err
}
