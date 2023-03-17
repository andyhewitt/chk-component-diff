# chk-component-diff

client-go to check component diff between clusters.

## To check component image diff

```bash
go run main.go -c=jpe2-caas1-prod4,jpe2-caas1-idprod1 -r=deploy -n=kube-system
```

## To check node label diff

Currently, this method will pick a random node which matches the label and compare.

```bash
go run main.go -c=jpe2-caas1-prod4,jpe2-caas1-idprod1 -l=cluster.aps.cpd.rakuten.com/noderole=master
```
