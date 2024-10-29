import json
import os
import sys
import re
import subprocess
import glob
# Define directories
SCHEMA_DIR = "schemas"
MODEL_DIR = "models"
PROTO_DIR = "proto"
SERVICE_DIR = "services"

# Map JSON schema types to Go types
type_mapping = {
    "string": "string",
    "integer": "uint64",
    "boolean": "bool",
    "number": "float64",
    "date-time": "time.Time"
}

def convert_field_name(field):
    parts = field.split('_')
    # Capitalize each part and join them to form camel case
    camel_cased = ''.join(part.capitalize() for part in parts)
    
    # Special handling for "id" to make it "ID"
    if camel_cased == 'Id':
        return 'ID'
    
    return camel_cased

def load_schema(schema_name):
    with open(f"{SCHEMA_DIR}/{schema_name}_schema.json") as f:
        return json.load(f)

def generate_go_model(schema_name, schema):
    model_name = convert_field_name(schema_name)
    properties = schema["properties"]
    required_fields = schema.get("required", [])

    model_lines = ["package models\n"]

    imports = set(["time"])

    custom_types = {}

    # Ensure created_at, updated_at, and id are in the properties
    if "created_at" not in properties:
        properties["created_at"] = {"type": "string", "format": "date-time"}
    if "updated_at" not in properties:
        properties["updated_at"] = {"type": "string", "format": "date-time"}
    if "id" not in properties:
        properties["id"] = {"type": "integer", "unique": True, "primary-key": True}

    # Type mapping dictionary
    type_mapping = {
        "string": "string",
        "integer": "uint64",
        "number": "float64",
        "boolean": "bool",
        "date-time": "time.Time",
        # Add other types as necessary
    }

    # Start building the model code
    model_lines.append("import (\n")
    for imp in sorted(imports):
        model_lines.append(f'\t"{imp}"\n')
    model_lines.append(")\n\n")

    # Start struct definition
    model_lines.append(f"type {model_name} struct {{\n")

    # Now process each field
    for field, specs in properties.items():
        field_type = type_mapping.get(specs["type"], "interface{}")
        if "format" in specs and specs["format"] == "date-time":
            field_type = type_mapping["date-time"]
        elif specs["type"] == "array" and "items" in specs:
            # Determine the type of array elements
            item_type = specs["items"]["type"]
            if item_type == "string" and specs.get("uniqueItems", False):
                # Generate custom type
                custom_type_name = convert_field_name(field)
                field_type = custom_type_name
                if custom_type_name not in custom_types:
                    custom_types[custom_type_name] = specs
                    # Add necessary imports for custom types
                    imports.update(["database/sql/driver", "encoding/json", "fmt"])
            else:
                field_type = f"[]{type_mapping.get(item_type, 'string')}"

        go_field_name = convert_field_name(field)
        tags = [f'json:"{field}"']

        # Add GORM and BSON tags
        gorm_tags = []
        bson_tag = f'{field}'
        if field in required_fields:
            gorm_tags.append('not null')
        if field == "id":
            gorm_tags.append('primaryKey')
            bson_tag = '_id'
        elif field == "created_at":
            gorm_tags.append('autoCreateTime;<-:create')
        elif field == "updated_at":
            gorm_tags.append('autoUpdateTime')
        else:
            if "[]" in field_type or field in custom_types:
                gorm_tags.append('type:json')

        if gorm_tags:
            tags.append(f'gorm:"{";".join(gorm_tags)}"')
        tags.append(f'bson:"{bson_tag}"')

        # Build validation tags
        validation_tags = []
        if "minLength" in specs:
            validation_tags.append(f"min={specs['minLength']}")
        if "maxLength" in specs:
            validation_tags.append(f"max={specs['maxLength']}")
        if "minimum" in specs:
            validation_tags.append(f"gte={specs['minimum']}")
        if specs.get("type") == "array":
            if specs.get("uniqueItems", False):
                validation_tags.append("unique")
            # Handle item-level validation
            item_specs = specs.get("items", {})
            if item_specs.get("type") == "string":
                item_validation = []
                if "minLength" in item_specs:
                    item_validation.append(f"min={item_specs['minLength']}")
                if "maxLength" in item_specs:
                    item_validation.append(f"max={item_specs['maxLength']}")
                if item_validation:
                    validation_tags.append(f"dive,{','.join(item_validation)}")

        if validation_tags:
            tags.append(f'validate:"{",".join(validation_tags)}"')

        # Combine tags
        tag_str = ' '.join(tags)

        # Append field definition to model_lines
        model_lines.append(f"\t{go_field_name} {field_type} `{tag_str}`\n")

    # Close struct definition
    model_lines.append("}\n")

    # Generate methods for custom types
    for type_name, specs in custom_types.items():
        model_lines.append(f"\n// {type_name} is a custom type for storing a slice of strings as JSON in the database.\n")
        model_lines.append(f"type {type_name} []string\n")

        model_lines.append(f"\n// Value converts the {type_name} to a JSON-encoded string for storing in the database.\n")
        model_lines.append(f"func (t {type_name}) Value() (driver.Value, error) {{\n")
        model_lines.append("\treturn json.Marshal(t)\n")
        model_lines.append("}\n")

        model_lines.append(f"\n// Scan converts a JSON-encoded value from the database into a {type_name} type.\n")
        model_lines.append("func (t *"+type_name+") Scan(value interface{}) error {\n")
        model_lines.append("\tb, ok := value.([]byte)\n")
        model_lines.append("\tif !ok {\n")
        model_lines.append("\t\treturn fmt.Errorf(\"failed to unmarshal JSON value: %v\", value)\n")
        model_lines.append("\t}\n")
        model_lines.append("\treturn json.Unmarshal(b, &t)\n")
        model_lines.append("}\n")

    # Generate GetID method
    model_lines.append(f"\nfunc (m *{model_name}) GetID() uint64 {{\n")
    model_lines.append("\treturn m.ID\n")
    model_lines.append("}\n")

    # Write the generated Go model to a file
    model_file_path = f"{MODEL_DIR}/{schema_name}.go"
    with open(model_file_path, "w") as f:
        f.writelines(model_lines)

    print(f"Generated Go model: {model_file_path}")

def generate_proto_file(schema_name, schema):
    model_name = convert_field_name(schema_name)
    properties = schema["properties"]
    required_fields = schema.get("required", [])

    # Initialize proto file lines
    proto_lines = [
        'syntax = "proto3";\n\n',
        "package proto;\n",
        f'option go_package = "./proto";\n\n',
    ]

    # Check if we need to import google.protobuf.Timestamp
    needs_timestamp_import = any(
        specs.get("format") == "date-time" for specs in properties.values()
    )
    if needs_timestamp_import:
        proto_lines.append('import "google/protobuf/timestamp.proto";\n\n')

    # Start the message definition
    proto_lines.append(f"message {model_name} {{\n")

    if "created_at" not in properties:
        properties["created_at"] = {"type": "string", "format": "date-time"}
    if "updated_at" not in properties:
        properties["updated_at"] = {"type": "string", "format": "date-time"}
    if "id" not in properties:
        properties["id"] = {"type": "integer", "unique": "true", 'primary-key':"true"}
    # Ensure required fields are processed first for better organization
    sorted_fields = sorted(properties.items(), key=lambda item: item[0] not in required_fields)

    # Generate fields for the proto message
    field_counter = 1
    for field, specs in sorted_fields:
        proto_type = "string"  # Default type
        if specs["type"] == "string":
            proto_type = "string"
        elif specs["type"] == "integer":
            proto_type = "uint64"
        elif specs["type"] == "boolean":
            proto_type = "bool"
        elif specs["type"] == "number":
            proto_type = "double"
        elif specs["type"] == "array" and "items" in specs:
            item_type = specs["items"]["type"]
            proto_type = f"repeated {type_mapping.get(item_type, 'string')}"
        if "format" in specs and specs["format"] == "date-time":
            proto_type = "google.protobuf.Timestamp"  # Use timestamp for date-time fields

        # Convert field names to Go-style camel case
        go_field_name = convert_field_name(field)

        # Add the field to the proto message
        proto_lines.append(f"    {proto_type} {go_field_name} = {field_counter};\n")
        field_counter += 1

    proto_lines.append("}\n\n")

    # Define service methods for CRUD operations
    proto_lines += [
        f"message Create{model_name}Request {{\n    {model_name} {schema_name} = 1;\n}}\n",
        f"message Create{model_name}Response {{\n    uint64 id = 1;\n    string message = 2;\n}}\n",
        f"message Get{model_name}Request {{\n    uint64 id = 1;\n}}\n",
        f"message Get{model_name}Response {{\n    {model_name} {schema_name} = 1;\n}}\n",
        f"message Update{model_name}Request {{\n    {model_name} {schema_name} = 1;\n}}\n",
        f"message Update{model_name}Response {{\n    string message = 1;\n}}\n",
        f"message Delete{model_name}Request {{\n    uint64 id = 1;\n}}\n",
        f"message Delete{model_name}Response {{\n    string message = 1;\n}}\n",
        f"service {model_name}Service {{\n",
        f"    rpc Create{model_name}(Create{model_name}Request) returns (Create{model_name}Response);\n",
        f"    rpc Get{model_name}(Get{model_name}Request) returns (Get{model_name}Response);\n",
        f"    rpc Update{model_name}(Update{model_name}Request) returns (Update{model_name}Response);\n",
        f"    rpc Delete{model_name}(Delete{model_name}Request) returns (Delete{model_name}Response);\n",
        "}\n"
    ]

    # Write the generated proto file
    proto_file_path = f"{PROTO_DIR}/{schema_name}.proto"
    with open(proto_file_path, "w") as f:
        f.writelines(proto_lines)

    print(f"Generated gRPC proto file: {proto_file_path}")
    print("To generate the gRPC code, run:")
    print(f"protoc --go_out=. --go-grpc_out=. {proto_file_path}")

def generate_service_impl(schema_name, schema):
    model_name = convert_field_name(schema_name)
    service_name = f"{model_name}ServiceServerImpl"
    service_file_path = f"{SERVICE_DIR}/{schema_name}_service_impl.go"

    service_lines = [
        f"package services\n\n",
        f'import (\n',
        f'    "fmt"\n',
        f'    "time"\n',
        f'    "context"\n',
        f'    "persistence-layer/models"\n',
        f'    "persistence-layer/orm"\n',
        f'    "persistence-layer/proto"\n',
        f'    "persistence-layer/utils"\n',
        f'    "google.golang.org/grpc"\n',
        f')\n\n',
        f'type {service_name} struct {{\n',
        f'    proto.Unimplemented{model_name}ServiceServer\n',
        f'    orm *orm.ORM\n',
        f'}}\n\n',
        f'func New{service_name}(orm *orm.ORM) *{service_name} {{\n',
        f'    return &{service_name}{{\n',
        f'        orm: orm,\n',
        f'    }}\n',
        f'}}\n\n',
    ]

    # Implement Create
    service_lines += [
        f'func (s *{service_name}) Create{model_name}(ctx context.Context, req *proto.Create{model_name}Request) (*proto.Create{model_name}Response, error) {{\n',
        f'    {schema_name} := models.{model_name}{{\n',
    ]

    # Map fields from proto request to Go model, including conversion for timestamps
    for field, specs in schema["properties"].items():
        go_field_name = convert_field_name(field)
        if specs.get("format") == "date-time":
            service_lines.append(f'        {go_field_name}: utils.ToTime(req.{model_name}.{go_field_name}),\n')
        else:
            service_lines.append(f'        {go_field_name}: req.{model_name}.{go_field_name},\n')

    service_lines += [
        f'    }}\n\n',
        f'    err := s.orm.Create(&{schema_name})\n',
        f'    if err != nil {{\n',
        f'        return nil, utils.HandleSQLError(err)\n',
        f'    }}\n\n',
        f'    return &proto.Create{model_name}Response{{\n',
        f'        Id: uint64({schema_name}.ID),\n',
        f'        Message: "{model_name} created successfully",\n',
        f'    }}, nil\n',
        f'}}\n\n'
    ]

    # Implement Get
    service_lines += [
        f'func (s *{service_name}) Get{model_name}(ctx context.Context, req *proto.Get{model_name}Request) (*proto.Get{model_name}Response, error) {{\n',
        f'    var {schema_name} models.{model_name}\n',
        f'    cacheKey := fmt.Sprintf("{schema_name}:%d", uint(req.Id))\n',
        f'    // Attempt to retrieve {schema_name} from cache\n',
        f'    err := s.orm.GetCache(cacheKey, &{schema_name})\n',
        f'    fromDb := false\n',
        f'    if err != nil || {schema_name}.ID == 0 {{\n',
        f'        // If {schema_name} is not found in cache, fetch from SQL database\n',
        f'        err := s.orm.Read(uint(req.Id), &{schema_name})\n',
        f'        if err != nil {{\n',
        f'            return nil, utils.HandleSQLError(err)\n',
        f'        }}\n',
        f'        fromDb = true\n',
        f'    }}\n\n',
        f'    // Cache the {schema_name} data with a TTL of 10 minutes\n',
        f'    if fromDb {{\n',
        f'        _ = s.orm.SetCache(cacheKey, &{schema_name}, 10*time.Minute)\n',
        f'    }}\n',
        f'    return &proto.Get{model_name}Response{{\n',
        f'        {model_name}: &proto.{model_name}{{\n',
    ]
    # Map fields from Go model to proto response, including conversion for timestamps
    for field, specs in schema["properties"].items():
        go_field_name = convert_field_name(field)
        if specs.get("format") == "date-time":
            service_lines.append(f'            {go_field_name}: utils.ToTimestamp({schema_name}.{go_field_name}),\n')
        else:
            service_lines.append(f'            {go_field_name}: {schema_name}.{go_field_name},\n')

    service_lines += [
        f'        }},\n',
        f'    }}, nil\n',
        f'}}\n\n'
    ]

    # Implement Update
    service_lines += [
        f'func (s *{service_name}) Update{model_name}(ctx context.Context, req *proto.Update{model_name}Request) (*proto.Update{model_name}Response, error) {{\n',
        f'    {schema_name} := models.{model_name}{{\n',
    ]

    # Map fields from proto request to Go model for update, including conversion for timestamps
    for field, specs in schema["properties"].items():
        go_field_name = convert_field_name(field)
        if specs.get("format") == "date-time":
            service_lines.append(f'        {go_field_name}: utils.ToTime(req.{model_name}.{go_field_name}),\n')
        else:
            service_lines.append(f'        {go_field_name}: req.{model_name}.{go_field_name},\n')

    service_lines += [
        f'    }}\n\n',
        f'    err := s.orm.Update(&{schema_name})\n',
        f'    if err != nil {{\n',
        f'        return nil, utils.HandleSQLError(err)\n',
        f'    }}\n\n',
        f'    cacheKey := fmt.Sprintf("{schema_name}:%d", uint(req.{model_name}.ID))\n',
        f'    _ = s.orm.SetCache(cacheKey, &{schema_name}, 10*time.Minute)\n',
        f'    \n\n',
        f'    return &proto.Update{model_name}Response{{\n',
        f'        Message: "{model_name} updated successfully",\n',
        f'    }}, nil\n',
        f'}}\n\n'
    ]

    # Implement Delete
    service_lines += [
        f'func (s *{service_name}) Delete{model_name}(ctx context.Context, req *proto.Delete{model_name}Request) (*proto.Delete{model_name}Response, error) {{\n',
        f'    err := s.orm.Delete(uint(req.Id), &models.{model_name}{{}})\n',
        f'    if err != nil {{\n',
        f'        return nil, utils.HandleSQLError(err)\n',
        f'    }}\n\n',
        f'    cacheKey := fmt.Sprintf("product:%d", uint(req.Id))\n'
        f'    _ = s.orm.DeleteCache(cacheKey)\n\n'
        f'    return &proto.Delete{model_name}Response{{\n',
        f'        Message: "{model_name} deleted successfully",\n',
        f'    }}, nil\n',
        f'}}\n\n'
    ]
    
    service_lines += [
        f'func (s *{service_name}) Register(server *grpc.Server) {{\n',
        f'    proto.Register{convert_field_name(schema_name)}ServiceServer(server, s)\n',
        f'}}\n\n'
    ]
    # Write the service implementation to a file
    with open(service_file_path, "w") as f:
        f.writelines(service_lines)

    print(f"Generated gRPC service implementation: {service_file_path}")


def main(schema_name):
    # Load the JSON schema
    schema = load_schema(schema_name)

    # Generate Go model based on the schema
    generate_go_model(schema_name, schema)

    # Generate gRPC proto file based on the schema
    generate_proto_file(schema_name, schema)

    # Generate the gRPC service implementation based on the schema
    generate_service_impl(schema_name, schema)

if __name__ == "__main__":
    if len(sys.argv) != 2:
        # No specific schema provided, process all schemas
        schema_pattern = re.compile(r'^(.+)_schema\.json$')
        for filename in os.listdir('schemas/'):  # Assuming the current directory, adjust if necessary
            match = schema_pattern.match(filename)
            print(match)
            if match:
                main(match.group(1))
    else:
        schema_name = sys.argv[1]
        main(schema_name)
    # Run the protoc command before running `update_main_file.py`
    try:
        # Find all .proto files in the 'proto' directory
        proto_files = glob.glob("proto/*.proto")
        if proto_files:
            subprocess.run(
                ["protoc", "--go_out=.", "--go-grpc_out=.", *proto_files],
                check=True
            )
            print("Successfully generated Go code from proto files.")
        else:
            print("No .proto files found in the 'proto' directory.")
            sys.exit(1)  # Exit if no proto files are found

    except subprocess.CalledProcessError as e:
        print(f"An error occurred while running protoc: {e}")
        sys.exit(1)  # Exit if protoc fails to prevent running further steps
    try:
        # This assumes `update_main_file.py` is in the same directory as the current script.
        subprocess.run(["python", "update_main_file.py"], check=True)
        print("Successfully ran update_main_file.py.")
    except subprocess.CalledProcessError as e:
        print(f"An error occurred while running update_main_file.py: {e}")

