# branch-cleaning
GitHubEnterpriseのリポジトリを指定し、`protectBranches`に設定したブランチ以外を削除するのみ。

## Usage

```
go run cleaning.go --repository {repository_name}
```

## Compile

```
GOOS=linux GOARCH=amd64 go build cleaning.go
GOOS=darwin GOARCH=amd64 go build cleaning.go
GOOS=windows GOARCH=amd64 go build cleaning.go
```