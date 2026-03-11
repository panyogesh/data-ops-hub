import json
import base64
import os
import glob

# 1. Load the configuration JSON
# Note: Ensure your JSON is valid (added missing comma after project_id)
config_path = 'config.json' 
with open(config_path, 'r') as f:
    config = json.load(f)

project_id = config.get("project_id")
sa_path = config.get("service_account_json")

# 2. Read and Base64 encode the service account file
with open(sa_path, 'rb') as sa_file:
    # Read bytes and encode to base64 string (equivalent to base64 -w0)
    sa_base64 = base64.b64encode(sa_file.read()).decode('utf-8')

# 3. Process all YAML files in the current folder
# Adjust path if your YAMLs are in a different directory
yaml_files = glob.glob("*.yaml") + glob.glob("*.yml")

for file_path in yaml_files:
    with open(file_path, 'r') as f:
        content = f.read()

    # Replace the placeholders
    new_content = content.replace("GCP_PROJECT_ID", project_id)
    new_content = new_content.replace("BASE64_OF_GCP_SERVICE_ACCOUNT_JSON", sa_base64)

    # Save the changes back to the file
    with open(file_path, 'w') as f:
        f.write(new_content)
    
    print(f"Updated: {file_path}")
