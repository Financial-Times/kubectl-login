#!/usr/bin/env bash
# Strict Mode - http://redsymbol.net/articles/unofficial-bash-strict-mode/
set -euo pipefail
IFS=$'\n\t'

#URL to download the kubeconfigs
BUCKET_URL=https://upp-kubeconfig.s3-eu-west-1.amazonaws.com

# Check if the bucket is reachable
status=$(curl -s --head -w %{http_code} $BUCKET_URL/index -o /dev/null)
if [ $status -ne "200" ] 
then
  echo "S3 bucket not reachable. Check your connection"
  exit 1
fi

# Get kubeconfigs
curl -s $BUCKET_URL/index > /tmp/index
for config in $(cat /tmp/index); do
curl -s $BUCKET_URL/$config > /tmp/$config
done

#Merge kubeconfigs
myarray=()
for i in $(cat /tmp/index); do
l=/tmp/$i
myarray+=($l)
done
KUBECONFIG=$(echo "${myarray[@]}" | sed 's/ /:/g') kubectl config view --merge=true --flatten=true > /tmp/kubeconfig



#Cleanup
for config in $(cat /tmp/index); do
rm -f /tmp/$config
done
rm -f /tmp/index

echo "New merged kubeconfig generated in /tmp/kubeconfig"
