#!/usr/bin/env bash

################################################################################

# Dioxy URL endpoint
DIOXY_URL="http://127.0.0.1:33407"

################################################################################

# error print message to /dev/stderr
#
# *: Message (String)
#
# Code: No
# Echo: Yes
error() {
  echo "$@" >/dev/stderr
}

# zbx.get returns ordinary item by given key
#
# 1: Device ID (String)
# 2: Metrics name (String)
#
# Code: No
# Echo: Yes
zbx.get() {
  local deviceId="$1"
  local metrics="$2"
  local value
  local retval

  body=$(curl -sL "$DIOXY_URL")
  retval=$?

  if [[ $retval -ne 0 ]] ; then
    error "Error: cannot retrieve metrics from dioxy server"
    echo ZBX_NOT_SUPPORTED
    exit
  fi

  keyFormat="${deviceId}@${metrics}"
  value=$(echo "$body" | jq -r '."'"${keyFormat}"'".value')
  retval=$?

  if [[ $retval -ne 0 ]] ; then
    error "Error: cannot parse metrics from received payload"
    echo ZBX_NOT_SUPPORTED
    exit
  fi

  echo "$value"
}

# zbx.discover returns list of available devices in Zabbix LLD format
#
# 1: Key (String)
#
# Code: No
# Echo: Yes
zbx.discover() {
  local retval

  body=$(curl -sL "$DIOXY_URL")
  retval=$?

  if [[ $retval -ne 0 ]] ; then
    error "Error: cannot retrieve metrics from dioxy server"
    echo ZBX_NOT_SUPPORTED
    exit
  fi

  devices=$(echo "$body" | jq -r '. | keys[]')

  IFS=$'\n'
  SEP=""

  JSON="{\"data\":["
  for d in $devices ; do
      deviceId=$(echo "$d" | awk -F '@' '{print $1}')
      JSON=$JSON"$SEP{\"{#DEVICE_ID}\":\"$deviceId\"}"
      SEP=", "
  done
  JSON=$JSON"]}"

  echo "$JSON"
}

# checkDeps checks if dependencies are installed
# Otherwise it finishes the program with an error (non-zero status)
#
# Code: No
# Echo: Yes
checkDeps() {
  for p in jq curl ; do
    if ! type "$p" &>/dev/null ; then
      echo "Error: unable to find: ${p}"
      exit 1
    fi
  done
}

################################################################################

# Main function
main() {
  local action="$1"
  shift

  case "$action" in
    "get")
      zbx.get "$@"
    ;;
    "discover")
      zbx.discover
    ;;
    *)
      error "Error: unsupported option"
      echo ZBX_NOT_SUPPORTED
    ;;
  esac
}

################################################################################

main "$@"
