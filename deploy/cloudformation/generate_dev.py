import os

prod_cft = 'elastic-agent-ec2.yml'
dev_cft = 'elastic-agent-ec2-dev.yml'

def edit_artifact_url(content):
  prodUrl = 'https://artifacts.elastic.co/downloads/beats/elastic-agent/'
  
  # TODO: Dynamically get the latest snapshot URL
  devUrl = 'https://snapshots.elastic.co/8.8.0-4c45f51b/downloads/beats/elastic-agent/'
  return content.replace(prodUrl, devUrl)

def main():
    script_path = os.path.abspath(__file__)
    current_dir = os.path.dirname(script_path)

    input_path = os.path.join(current_dir, prod_cft)
    output_path = os.path.join(current_dir, dev_cft)

    with open(input_path, 'r') as f:
        file_contents = f.read()

    modified_contents = edit_artifact_url(file_contents)

    with open(output_path, 'w') as f:
        f.write(modified_contents)

    print(f'Created {output_path}')

main()
