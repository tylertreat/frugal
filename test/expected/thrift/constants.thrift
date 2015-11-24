
typedef string System

const SYSTEM_FOO = "foo"
const SYSTEM_BAR = "bar"
const SYSTEM_BAZ = "baz"

struct SchemaProperty {
    1:string id,
    2:string title,
    3:bool inScope,
}

struct Schema {
    1:string foo,
    2:list<SchemaProperty> properties,
}

const map<i64,string> SIMPLE_MAP_INT_KEYS = {0: "foo", 1: "bar", 2: "baz"}
const map<string,i64> SIMPLE_MAP_STRING_KEYS = {"foo": 0, "bar": 1, "baz": 2}

const map<System,Schema> REGISTERED_SCHEMAS = {
    SYSTEM_FOO: {
        "name": SYSTEM_FOO,
        "properties": [
            {
                "id": "prop1",
                "title": "property 1",
                "inScope": true,
            },
            {
                "id": "prop2",
                "title": "property 2",
            },
        ],
    },

    "bar": {
        "name": SYSTEM_BAR,
        "properties": [
            {
                "id": "prop3",
                "title": "property 3",
                "inScope": true,
            },
            {
                "id": "prop4",
                "title": "property 4",
            },
        ],
    },
}
