# Install CSPM integration on a cloud deployment with standalone docker agent

## Prerequiste
1. Docker
2. Connect to AWS account using the CLI

## 1. Deploy Cloud Environment
- Go to https://cloud.elastic.co/home
- Log into with your elastic account
- Select `Create Deployment`
- Adjust the settings for your need, note that certain versions only exist on certain regions. 
- At the time of writing this, `latest` and `snapshot` versions are avilable on `us-west2` 

<img width="900" alt="image" src="https://user-images.githubusercontent.com/51442161/222231445-5033bd97-9f19-4241-9784-876e92417a23.png">

Launch your Cloud Deployment

## 2. Optional Step - Verify Cloud Deployment Commit
In order to confirm that the commit SHA of your deployment is matching the commit SHA of in the DRA
- In your deployment, navigate to `/app/status`, the commit of your Kibana version will be displayed at the top

<img width="900" alt="image" src="https://user-images.githubusercontent.com/51442161/222398648-348cae1d-2a5e-4039-aa3b-4ce9983d3b04.png">

- Navigate to https://artifacts-staging.elastic.co/dra-info/index.html, here you can see all the latest commit for all snapshot and staging versions
- Click on `JSON report` next to the version you are checking

<img width="900" alt="image" src="https://user-images.githubusercontent.com/51442161/222493787-c2a3c2ae-72d4-44e6-9a40-ef7334090c44.png">

- A new tab should open, navigate to the `summary url` provided

```
{
  "version" : "8.7.0",
  "build_id" : "8.7.0-046d305b",
  "manifest_url" : "https://staging.elastic.co/8.7.0-046d305b/manifest-8.7.0.json",
  "summary_url" : "https://staging.elastic.co/8.7.0-046d305b/summary-8.7.0.html" // <-- Navigate to this URL
}
```

- In the Elastic Stack Release page, click on `kibana`

<img width="900" alt="image" src="https://user-images.githubusercontent.com/51442161/222495735-dea09020-3e08-45d9-8d2d-4edc6a9d5ee7.png">

- The commit SHA displayed should match the commit SHA from your deployment status that we checked at the first stage

<img width="900" alt="image" src="https://user-images.githubusercontent.com/51442161/222495899-7164a25f-b8e6-4970-8b83-a39a3e1d094a.png">

## 3. Create Agent Policy
- Navigate to Fleet > Agent policies and select `Create agent policy`

<img width="900" alt="image" src="https://user-images.githubusercontent.com/51442161/222233724-fcdc7d5d-d35b-4fb8-aed9-c157471c789d.png">

- In the Flyout, give it a name (you'll need it later) and click on `Create agent policy`
- A new policy has been added to your list of agent policies, click on its name and then on `Add integration`
- In the intergations screen, select the needed integration. For this tutorial, we will use CSPM

<img width="900" alt="image" src="https://user-images.githubusercontent.com/51442161/222234500-39dbda6c-8880-4f43-99f4-9cc2d3c7a655.png">

- Click on `Add Cloud Security Posture Management (CSPM)`
- In the configurations screen, select the Amazon Web Services provider and add your `Access Key ID` and `Secret Access Key` in the `Direct access keys` option.

- Run `cat ~/.aws/credentials` to get the keys

<img width="900" alt="image" src="https://user-images.githubusercontent.com/51442161/222235991-5ea07776-d5fe-4a9c-9629-a619c5326970.png">

Save the intergation

## 4. Add Standalone Docker Agent
- Now that you have an agent policy with a configured CSPM intergation navigate to Fleet>Agents and select `Add Agent`
- Select your CSPM intergation from the drop down list under `What type of host are you adding?`
- On step 3 of the Flyout, you will be provided with the setup command for the agent, for example:

```
curl -L -O https://artifacts.elastic.co/downloads/beats/elastic-agent/elastic-agent-8.7.0-SNAPSHOT-darwin-x86_64.tar.gz
tar xzvf elastic-agent-8.7.0-SNAPSHOT-darwin-x86_64.tar.gz
cd elastic-agent-8.7.0-SNAPSHOT-darwin-x86_64
sudo ./elastic-agent install --url=https://901181905d2049f98455066cda0e6717.fleet.us-west2.gcp.elastic-cloud.com:443 --enrollment-token=cUV0Mm5vWUJlaUVHQ3hJWTJqOXQ6bEJnZUNmWGFRWkMzR3BHeEFRS2dYZw==
```

- Since we want to deploy the agent in docker we will need to run a different command.
- Extract the `fleet-server-host-url`(--url) and `enrollment-token`(--enrollment-token) values from the provided command and use them in the following command instead:

```
docker run -d --platform=linux/x86_64 \
-e "FLEET_URL=<fleet-server-host-url>" \
-e "FLEET_ENROLLMENT_TOKEN=<enrollment-token>" \
-e "FLEET_ENROLL=1" \
docker.elastic.co/beats/elastic-agent:8.7.0-SNAPSHOT
```
> **Note**
> make sure to also change the version according to what you are using)

At this point you should see your agent working in your docker and under `Confirm agent enrollment` section of the Flyout.

## 5. Discover
- Go to Discover and make sure you have docs in the `findings` and `findings-latest` indecies, `findings-latest` can take a few minutes to populate.
- If you do, congrats you have a working CSPM cloud deployment.
