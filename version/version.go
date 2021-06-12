package version

var (
	Version    string
	GitCommit  string
	DevVersion = "dev"
)

func BuildVersion() string {
	if len(Version) == 0 {
		return DevVersion
	}
	return Version
}

func GetReleaseInfo() (sha, release string) {
	return GitCommit, BuildVersion()
}
