# Cluster Image and Label Checker

This project is designed to check whether two or more clusters' core components are using the same image URL or whether two or more clusters' nodes have identical labels.

## Installation

To install this project, please follow the steps below:

1. Clone the repository to your local machine
2. Make sure you have GoLang installed on your machine
3. Navigate to the project directory
4. Run `go build` to build the project
5. Run `./chk-component-diff [flag]` to run the project

## Usage

The following flags are available for use with this project:

- **-c:** Specify which clusters to check, comma-separated
- **-n:** Specify which namespace to check, comma-separated
- **-r:** Specify which resources to check (deployment, statefulset, etc.), comma-separated
- **-table-length:** Specify the cell width when output is printed
- **-caas:** If this flag is set, it will only check CaaS core components inside namespace *kube-system*, *caas-system*, *caas-csi* and *istio-system*
- **-l:** If this flag is set, it will pick one node which matches this label in one of each clusters and compare its labels.

To run the project with flags, use the following format:

```bash
./project-name -c cluster1,cluster2 -n namespace1,namespace2 -r deployment,statefulset -table-length 20
```
