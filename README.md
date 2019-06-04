# ESQL: Translate SQL to Elasticsearch DSL

## Milestones

### M1
- [x] comparison operators: =, !=, <, >, <=, >=
- [x] boolean operators: AND, OR, NOT
- [x] parenthesis: ()
- [x] auto testing
- [x] setup git branch for pull request and code review
- [x] keyword: LIMIT, SIZE
- [x] depedencies management and golint checking

### M2
- [x] keyword: IS NULL, IS NOT NULL (missing check)
- [x] keyword: BETWEEN
- [x] keyword: IN, NOT IN
- [ ] keyword: LIKE, NOT LIKE

### M3
- [ ] select specific columns
- [ ] keyword: ORDER BY
- [ ] keyword: GROUP BY

### M4
- [ ] special handling: ExecutionTime field
- [ ] key whitelist filtering
- [ ] column name filtering
- [ ] pagination, search after

### Misc
- [ ] optimization: docvalue_fields, term&keyword
- [ ] documentation
- [ ] test cases for unsupported and invalid queries


## Motivation
Currently we are using [elasticsql](https://github.com/cch123/elasticsql). However it only support up to ES V2.x while [Cadence](https://github.com/uber/cadence) is using ES V6.x. Beyond that, Cadence has some specific requirements that not supported by elasticsql yet.

Current Cadence query request processing steps are listed below:
- generate SQL from query
- use elasticsql to translate SQL to DSL
- ES V6.x does not support "missing" field, convert "missing" to "bool","must_not","exist" for ExecutionTime query if any
- complete "range" field for ExecutionTime query by adding {"gt": 0}
- add domain query
- key whitelist filtering
- delete some useless field like "from", "size"
- modify sorting field
- setup search after for pagination

This project is based on [elasticsql](https://github.com/cch123/elasticsql) and aims at dealing all these addtional processing steps and providing an api to generate DSL in one step for visibility usage in Cadence.


## Testing Module
We are using elasticsearch's SQL translate API as a reference in testing. Testing contains 3 basic steps:
- using elasticsearch's SQL translate API to translate sql to dsl
- using our library to convert sql to dsl
- query local elasticsearch server with both dsls, check the results are identical

There are some specific features not covered in testing yet:
- `LIMIT` keyword: when order is not specified, identical queries with LIMIT can return different results
- `LIKE` keyword: ES V6.5's sql api does not support regex search but only wildcard (only support shell wildcard `%` and `_`)

Testing steps:
- download elasticsearch v6.5 (optional: kibana v6.5) and unzip
- run `chmod u+x start_service.sh test_all.sh`
- run `./start_service.sh <elasticsearch_path> <kibana_path>` to start a local elasticsearch server
- optional: modify `sqls.txt` to add custom SQL queries as test cases
- optional: run `python gen_test_date.py -dcmi <number of documents> <missingRate>` to customize testing data set
- run `./test_all.sh` to run all the test cases
- generated dsls are stored in `dsls.txt` and `dslsPretty.txt` for reference


## esql vs elasticsql
|Item|esql|elasticsql|
|:-:|:-:|:-:|
|scoring|using "filter" to avoid scoring analysis and save time|using "must" which calculates scores|
|missing check|support IS NULL, IS NOT NULL|does not support IS NULL, using colName = missing which is not standard sql|
|NOT expression|support NOT, convert NOT recursively since elasticsearch's must_not is not the same as boolean operator NOT in sql|not supported|
|LIKE expression|using "regexp", support standard regex syntax|using "match_phrase", only support '%' and the smallest match unit is space separated word|
|optimization|no redundant {"bool": {"filter": xxx}} wrapped|all queries wrapped by {"bool": {"filter": xxx}}|

## ES V2.x vs ES V6.5
|Item|ES V2.x|ES v6.5|
|:-:|:-:|:-:|
|missing check|{"missing": {"field": "xxx"}}|{"must_not": {"exist": {"field": "xxx"}}}|