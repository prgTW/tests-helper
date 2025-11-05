package junit

import "encoding/xml"

// TestSuite represents a JUnit XML test suite element.
type TestSuite struct {
	XMLName    xml.Name    `xml:"testsuite"`
	Name       string      `xml:"name,attr"`
	File       string      `xml:"file,attr"`
	Time       string      `xml:"time,attr"`
	TestSuites []TestSuite `xml:"testsuite"`
}

// TestSuites represents the root element of JUnit XML.
type TestSuites struct {
	XMLName    xml.Name    `xml:"testsuites"`
	TestSuites []TestSuite `xml:"testsuite"`
}

// Test represents a single test with its execution time.
type Test struct {
	Name string
	Time float64
}
