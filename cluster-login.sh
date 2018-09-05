#!/bin/bash
output=$(./kubectl-login $1 | tee /dev/stderr)

kubectlLoginOutput=($output)

if [ ${#kubectlLoginOutput[@]} -eq 2  ]
then
    export KUBECONFIG=${output} | awk '{print $2; }'
    echo "Logged in to $1. Using KUBECONFIG=$KUBECONFIG"
elif [ ${#kubectlLoginOutput[@]} -eq 1 ]
then
    export KUBECONFIG=${output}
    echo "Logged in to $1. Using KUBECONFIG=$KUBECONFIG"
else
    echo "$output"
fi
