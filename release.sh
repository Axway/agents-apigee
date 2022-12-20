#!/bin/bash
shopt -s nocasematch

# you can add variables like this to test locally. Just run ./release.sh. Note that to run on MAC, you must install bash 5.x. 
# Then, to run the script you must do: /usr/local/bin/bash ./release.sh
# TEAMS_WEBHOOK_URL="foo.bar"
# TAG="1.2.3"

# Would like to use this, but can't get it to work in curl command
#export COMMON_CURL_HEADER=`printf -- '-H "PRIVATE-TOKEN:${GIT_API_TOKEN}" -H "Accept:application/json" -H "Content-Type:application/json"'`
export H_ACCEPT="Accept:application/json"
export H_CONTENT="Content-Type:application/json"
export H_TOKEN="PRIVATE-TOKEN:${GIT_API_TOKEN}"

# validate all of the required variables
check_required_variables() {
    echo "Validating the required environment variables..."

    [ -z "${TEAMS_WEBHOOK_URL}" ] && echo "TEAMS_WEBHOOK_URL variable not set" && exit 1
    [ -z "${TAG}" ] && echo "TAG variable not set" && exit 1
    [ -z "${SDK}" ] && echo "SDK variable not set" && exit 1

    pat='[0-9]+\.[0-9]+\.[0-9]'
    if [[ ! ${TAG} =~ $pat ]]; then
        echo "TAG variable must be of the form X.X.X"
        exit 1
    fi

    return 0
}

get_sdk_version()
{
    # pull out the SDK version from go.mod
    ver=$(grep 'github.com/Axway/agent-sdk v' ./discovery/go.mod)

    # Set space as the delimiter
    IFS=' '

    # Read the split words into an array
    read -a strarr <<< $ver
    export SDK=${strarr[1]}
}

post_to_teams() {
    rel_date=$(date +'%m/%d/%Y')
    JSON="{
        \"@type\": \"MessageCard\",
        \"@context\": \"http://schema.org/extensions\",
        \"summary\": \"Agent Release Info\",
         \"sections\": [{
             \"facts\": [{
                 \"name\": \"Date:\",
                 \"value\": \"${rel_date}\"
                 }, {
                 \"name\": \"Info:\",
                 \"value\": \"${1}\"
             }]
         }]
        }"
    curl -v ${TEAMS_WEBHOOK_URL} \
    -H 'Content-Type: application/json' \
    -d "${JSON}" &> /dev/null
}

main() {
    # validate required variables
    get_sdk_version
    check_required_variables

    if [ $? -eq 1 ]; then
        echo "No release info being generated."
        exit 1
    fi

    # gather stats
    releaseStats="- SDK version: ${SDK}\n"
    releaseStats+="- Apigee agents version: ${TAG}\n"

    echo -e "Full Release Info:\n"${releaseStats}
    post_to_teams "${releaseStats}"
    exit 0
}

main $@
