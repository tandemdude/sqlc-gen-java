set -e

current_version=$(grep -oP 'const PluginVersion = "\K[0-9]+\.[0-9]+\.[0-9]+' internal/core/constants.go)

if [[ -z "$current_version" ]]; then
  echo "Error: Unable to find PluginVersion"
  exit 1
fi

IFS='.' read -r MAJOR MINOR PATCH <<< "$current_version"

if [[ "$1" == "major" ]]; then
  ((MAJOR++))
  MINOR=0
  PATCH=0
elif [[ "$1" == "minor" ]]; then
  ((MINOR++))
  PATCH=0
elif [[ "$1" == "patch" ]]; then
  ((PATCH++))
else
  echo "Error: Invalid version type. Use 'major', 'minor', or 'patch'."
  exit 1
fi

new_version="$MAJOR.$MINOR.$PATCH"
echo "new plugin version is: $new_version"

sed -i "s/const PluginVersion = \"[0-9]\+\.[0-9]\+\.[0-9]\+\"/const PluginVersion = \"$new_version\"/" internal/core/constants.go

echo "VERSION=v$new_version" >> $GITHUB_OUTPUT
