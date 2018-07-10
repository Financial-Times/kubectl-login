#! /usr/local/bin/fish

set output (env KUBECONFIG=$KUBECONFIG kubectl-login $argv[1])
if test $status -eq 0
    set KUBECONFIG {$output}
    echo "Logged in to $argv[1]. Using KUBECONFIG=$KUBECONFIG"
else
    echo "$output"
end
