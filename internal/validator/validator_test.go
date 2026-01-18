package validator

import "testing"

func TestValidateCode_Safe(t *testing.T) {
	v := NewValidator()

	safeCode := `
import arcpy

layer = "Schools"
count = int(arcpy.management.GetCount(layer)[0])
arcpy.AddMessage(f"Count: {count}")
`

	result := v.ValidateCode(safeCode)

	if !result.IsValid {
		t.Errorf("Safe code marked as invalid: %v", result.Errors)
	}

	if result.Score < 80 {
		t.Errorf("Safe code scored too low: %d", result.Score)
	}
}

func TestValidateCode_FileRemoval(t *testing.T) {
	v := NewValidator()

	dangerousCode := `
import os
os.remove("/important/file.txt")
`

	result := v.ValidateCode(dangerousCode)

	if result.IsValid {
		t.Error("Dangerous file removal code marked as valid")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected errors for dangerous code")
	}
}

func TestValidateCode_SystemCall(t *testing.T) {
	v := NewValidator()

	dangerousCode := `
import subprocess
subprocess.call(["rm", "-rf", "/"])
`

	result := v.ValidateCode(dangerousCode)

	if result.IsValid {
		t.Error("System call code marked as valid")
	}
}

func TestValidateCode_NetworkOperation(t *testing.T) {
	v := NewValidator()

	dangerousCode := `
import urllib.request
urllib.request.urlopen("http://evil.com")
`

	result := v.ValidateCode(dangerousCode)

	if result.IsValid {
		t.Error("Network operation code marked as valid")
	}
}

func TestValidateCode_ArcPyBuffer(t *testing.T) {
	v := NewValidator()

	safeCode := `
import arcpy
arcpy.analysis.Buffer("Rivers", "river_buffer", "1000 Meters")
arcpy.AddMessage("Buffer created")
`

	result := v.ValidateCode(safeCode)

	if !result.IsValid {
		t.Errorf("Valid arcpy code marked as invalid: %v", result.Errors)
	}
}
