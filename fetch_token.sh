#!/bin/bash

prom_token () { 
    kubectl get secrets \
        -ojson \
        -n openshift-monitoring \
        --context "$1" | jq '
        [
            .items[] | 
            select(
                (.metadata.annotations | has("kubernetes.io/service-account.name")) 
                and 
                (.metadata.annotations["kubernetes.io/service-account.name"] == "prometheus-k8s") 
                and (.data | has("token"))
            ) 
        ] | .[0].data.token' -r | base64 --decode
}

context=""
print_usage() {
    printf "
        -c | cluster context to use:
    
"
}

while getopts "c:" f; do
    case "$f" in
    c)
        context=${OPTARG}
        ;;
    *)
        print_usage
        exit 1
        ;;
    esac
done

prom_token $context