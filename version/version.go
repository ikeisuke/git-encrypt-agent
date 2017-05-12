package version

import (
	"fmt"
)

type Version struct {
	Major                      int
	Minor                      int
	Patch                      int
	PreReleaseVersionIndentity string
	BuildMetadata              string
}

var CmdVersion = Version{
	Major: 0,
	Minor: 1,
	Patch: 0,
}

func (v Version) String() string {
	versionString := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if len(v.PreReleaseVersionIndentity) > 0 {
		versionString = fmt.Sprintf("%s-%s", versionString, v.PreReleaseVersionIndentity)
	}
	if len(v.BuildMetadata) > 0 {
		versionString = fmt.Sprintf("%s+%s", versionString, v.BuildMetadata)
	}
	return versionString
}
