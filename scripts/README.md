# Bootstrapping the Database

## Get source data

```bash
> mkdir data
> cd data
> wget https://fdc.nal.usda.gov/fdc-datasets/FoodData_Central_csv_2022-10-28.zip
> unzip FoodData_Central_csv_2022-10-28.zip
# TODO remove duplicates and fix badly formed rows
> gzip food.csv food_nutrient.csv nutrient.csv
```

## Load into the database

* Start Neo4j, with a volume mounted at `/var/lib/neo4j/import/`
  * The volume should contain the gzipped CSVs from above
* Run the queries contained in `cypher/import_usda.cypher`
