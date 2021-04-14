#!/bin/bash

#
# mkrpm.sh
#   --name {name}
#   --maintainer {name}
#   --epic {#}
#   --iteration {#}
#   --version {#}
#   --after-install {file}
#

POSITIONAL=()
while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    -n|--name)
    RPM_NAME="$2"
    shift # past argument
    shift # past value
    ;;
    -m|--maintainer)
    RPM_MAINTAINER="$2"
    shift # past argument
    shift # past value
    ;;
    -e|--epoch)
    RPM_EPOCH="$2"
    shift # past argument
    shift # past value
    ;;
    -i|--iteration)
    RPM_ITER="$2"
    shift # past argument
    shift # past value
    ;;
    -v|--version)
    RPM_VERSION="$2"
    shift # past argument
    shift # past value
    ;;
    -a|--after-install)
    RPM_AFTER_INSTALL="$2"
    shift # past argument
    shift # past value
    ;;
    -u|--url)
    RPM_URL="$2"
    shift # past argument
    shift # past value
    ;;
    -s|--summary)
    RPM_SUMMARY="$2"
    shift # past argument
    shift # past value
    ;;
    *)    # unknown option
    POSITIONAL+=("$1") # save it in an array for later
    shift # past argument
    ;;
esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

if [[ -z "$RPM_EPOCH" ]]; then
    RPM_EPOCH=1
fi

echo "RPM_NAME=$RPM_NAME"
echo "RPM_MAINTAINER=$RPM_MAINTAINER"
echo "RPM_EPOCH=$RPM_EPOCH"
echo "RPM_ITER=$RPM_ITER"
echo "RPM_VERSION=$RPM_VERSION"
echo "RPM_AFTER_INSTALL=$RPM_AFTER_INSTALL"
echo "RPM_URL=$RPM_URL"
echo "RPM_SUMMARY=$RPM_SUMMARY"

pwd && ls
cd .rpm/$RPM_NAME && fpm -s dir -t rpm \
	-n $RPM_NAME \
	-m $RPM_MAINTAINER \
	-p ../dist/ \
	--rpm-os linux \
	--verbose \
	--epoch $RPM_EPOCH \
	--iteration "$RPM_ITER" \
	--version "$RPM_VERSION" \
	--after-install "$RPM_AFTER_INSTALL" \
	--url "$RPM_URL" \
	--description "$RPM_SUMMARY" \
	.
