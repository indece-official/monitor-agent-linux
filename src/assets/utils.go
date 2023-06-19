// Copyright indece UG (haftungsbeschränkt) - All rights reserved.
//
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential
//
// Written by Stephan Lukas <stephan.lukas@indece.com>, 2022

package assets

import "io/ioutil"

// ReadFile reads a asseet file as string
func ReadFile(filename string) (string, error) {
	file, err := Assets.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
