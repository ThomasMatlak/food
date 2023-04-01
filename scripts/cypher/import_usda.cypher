CREATE CONSTRAINT id IF NOT EXISTS
FOR (r:Resource)
REQUIRE r.id IS UNIQUE;

CREATE TEXT INDEX food_id_idx IF NOT EXISTS
FOR (f:Food)
ON f.id;

CREATE FULLTEXT INDEX food_name_search_idx IF NOT EXISTS
FOR (f:Food)
ON EACH [f.name];

CREATE TEXT INDEX nutrient_id_idx IF NOT EXISTS
FOR (n:Nutrient)
ON n.id;

CREATE FULLTEXT INDEX ingredient_name_search_idx IF NOT EXISTS
FOR (n:Nutrient)
ON EACH [n.name];

:auto LOAD CSV WITH HEADERS FROM "file:///nutrient.csv.gz" AS row
CALL {
    WITH row
    MERGE (n:Nutrient:Resource {id: "grn:tm-food:nutrient:resource:" + row.id})
    ON CREATE SET n += {
        name: row.name,
        unit_name: row.unit_name
    }
} IN TRANSACTIONS;

:auto LOAD CSV WITH HEADERS FROM "file:///food.csv.gz" AS row
CALL {
    WITH row
    MERGE (f:Food:Resource {id: "grn:tm-food:food:resource:" + row.fdc_id})
    ON CREATE SET f += {
        name: row.description,
        // portions: [],
        created: row.publication_date
    }
} IN TRANSACTIONS;

:auto LOAD CSV WITH HEADERS FROM "file:///food_nutrient.csv.gz" AS row
CALL {
    WITH row
    // Don't why, but the db doesn't want to use the :Food or :Nutrient id indexes. It will use the :Resource id index though.
    MATCH(n:Nutrient:Resource {id: "grn:tm-food:nutrient:resource:" + row.nutrient_id})
    MATCH(f:Food:Resource {id: "grn:tm-food:food:resource:" + row.fdc_id})
    MERGE (f)-[rel:HAS_NUTRIENT]->(n)
    ON CREATE SET rel += {amount: toFloat(row.amount)}
} IN TRANSACTIONS;
