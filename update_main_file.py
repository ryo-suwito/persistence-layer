import os
import re

# Define directories and the target Go file to update
MODELS_DIR = "models"
SERVICES_DIR = "services"
TARGET_GO_FILE = "cmd/main.go"
MODEL_PATTERN = re.compile(r'type (\w+) struct')
SERVICE_PATTERN = re.compile(r'type (\w+ServiceServerImpl) struct')

def find_model_structs():
    """Scan the models directory for Go files and extract model struct names."""
    model_structs = []

    for file_name in os.listdir(MODELS_DIR):
        if file_name.endswith(".go"):
            file_path = os.path.join(MODELS_DIR, file_name)
            
            with open(file_path, 'r') as file:
                content = file.read()
                matches = MODEL_PATTERN.findall(content)
                model_structs.extend(matches)

    return model_structs

def find_service_implementations():
    """Scan the services directory for Go files and extract service implementation names."""
    service_implementations = []

    for file_name in os.listdir(SERVICES_DIR):
        if file_name.endswith(".go"):
            file_path = os.path.join(SERVICES_DIR, file_name)
            
            with open(file_path, 'r') as file:
                content = file.read()
                matches = SERVICE_PATTERN.findall(content)
                service_implementations.extend(matches)

    return service_implementations

def update_main_go_file(models, services):
    """Update the TARGET_GO_FILE with model auto-migrations and service implementations."""
    with open(TARGET_GO_FILE, 'r') as file:
        content = file.read()

    # Construct the new models array content for AutoMigrate
    model_instances = [
        "&models." + model + "{}," for model in models
    ]
    new_models_content = "\n        ".join(model_instances)

    # Define the replacement pattern for the AutoMigrate section
    new_auto_migrate_content = (
        "    // Run GORM auto-migration for your models here\n"
        "    db := sqlAdapter.GetDB()\n"
        "    err = db.AutoMigrate(\n"
        "        " + new_models_content + "\n"
        "    )\n"
        "    if err != nil {\n"
        "        log.Fatalf(\"Failed to auto migrate models: %v\", err)\n"
        "    }\n"
        "    log.Println(\"Auto migration completed successfully.\")"
    )

    # Replace the AutoMigrate block using regex
    content = re.sub(
        r"// Run GORM auto-migration for your models here(.|\s)*?log.Println\(\"Auto migration completed successfully\.\"\)",
        new_auto_migrate_content,
        content,
        flags=re.MULTILINE
    )

    # Construct the new services array content for GetAllServices
    service_instances = [
        "services.New" + service + "(ormLayer)," for service in services
    ]
    new_services_content = "\n        ".join(service_instances)

    # Define the replacement pattern for the GetAllServices function
    new_function_content = (
        "func GetAllServices(ormLayer *orm.ORM) []RegisterableService {\n"
        "    return []RegisterableService{\n"
        "        " + new_services_content + "\n"
        "    }"
    )

    # Replace the GetAllServices function content using regex
    content = re.sub(
        r"func GetAllServices\(.+?\) \[\]RegisterableService \{(.|\s)*?\}",
        new_function_content,
        content,
        flags=re.MULTILINE
    )

    # Write the updated content back to the file
    with open(TARGET_GO_FILE, 'w') as file:
        file.write(content)

    print(f"Updated {TARGET_GO_FILE} with the following models for auto-migration:")
    for model in models:
        print(f" - {model}")

    print("\nUpdated with the following services for registration:")
    for service in services:
        print(f" - {service}")

def main():
    # Step 1: Find all model structs in the models directory
    models = find_model_structs()

    # Step 2: Find all service implementations in the services directory
    services = find_service_implementations()

    # Step 3: Update the cmd/main.go file with detected models and services
    if models or services:
        update_main_go_file(models, services)
    else:
        print("No models or service implementations found.")

if __name__ == "__main__":
    main()
