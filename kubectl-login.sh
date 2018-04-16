#!/bin/bash
output=$(./kubectl-login $1)
if [ $? -eq 0  ]
then
    export KUBECONFIG=${output}
    echo "Logged in to $1."
else
    echo "$output"
fi
