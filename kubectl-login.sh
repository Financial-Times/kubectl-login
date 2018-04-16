#!/bin/bash
echo -n "Enter token: "
output=$(./kubectl-login $1)
if [ $? -eq 0  ]
then
    export KUBECONFIG=${output}
    echo "Logged in to $1."
else
    echo -e "\r$output"
fi
