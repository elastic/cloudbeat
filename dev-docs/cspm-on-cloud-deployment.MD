# Install CSPM integration on a cloud deployment with standalone docker agent

## Deploy Cloud Environment
Go to https://cloud.elastic.co/home
Log into with your elastic account
Click on `Create Deployment`
Adjust the settings for your need, note that certain versions only exist on certain regions. 
At the time of writing this, `latest` and `snapshot` versions are avilable on `us-west2` 
![image](https://user-images.githubusercontent.com/51442161/222231445-5033bd97-9f19-4241-9784-876e92417a23.png)

Launch your Cloud Deployment

## Optional Step - Verify Cloud Deployment Commit


## Create Agent Policy
Navigate to Fleet>Agent policies and click on `Create agent policy`
<img width="1254" alt="image" src="https://user-images.githubusercontent.com/51442161/222233724-fcdc7d5d-d35b-4fb8-aed9-c157471c789d.png">

In the Flyout, give it a name (you'll need it later) and click on `Create agent policy`
A new policy has been added to your list of agent policies, click on its name and then on `Add integration`
In the intergations screen, select the needed integration, for this tutorial, we will use CSPM
<img width="1703" alt="image" src="https://user-images.githubusercontent.com/51442161/222234500-39dbda6c-8880-4f43-99f4-9cc2d3c7a655.png">

Click on `Add Cloud Security Posture Management (CSPM)`
In the configurations screen, select the Amazon Web Services provider and add your `Access Key ID` and `Secret Access Key` in the `Direct access keys` option.

Run `cat ~/.aws/credentials` to get the keys
![image](https://user-images.githubusercontent.com/51442161/222235991-5ea07776-d5fe-4a9c-9629-a619c5326970.png)

Save the intergation

## Add Standalone Docker Agent
Now that you have an agent policy with a configured CSPM intergation navigate to Fleet>Agents and select `Add Agent`
Select your CSPM intergation from the drop down list under `What type of host are you adding?`
On step 3 of the Flyout, you will be provided with the setup command for the agent, for example:

```
curl -L -O https://artifacts.elastic.co/downloads/beats/elastic-agent/elastic-agent-8.7.0-SNAPSHOT-darwin-x86_64.tar.gz
tar xzvf elastic-agent-8.7.0-SNAPSHOT-darwin-x86_64.tar.gz
cd elastic-agent-8.7.0-SNAPSHOT-darwin-x86_64
sudo ./elastic-agent install --url=https://901181905d2049f98455066cda0e6717.fleet.us-west2.gcp.elastic-cloud.com:443 --enrollment-token=cUV0Mm5vWUJlaUVHQ3hJWTJqOXQ6bEJnZUNmWGFRWkMzR3BHeEFRS2dYZw==
```

Since we want to deploy the agent in docker we will need to run a different command.
Extract the `fleet-server-host-url` and `enrollment-token` values from the provided command and use them in the following command instead:

```
docker run -d --platform=linux/x86_64 \
-e "FLEET_URL=<fleet-server-host-url>" \
-e "FLEET_ENROLLMENT_TOKEN=<enrollment-token>" \
-e "FLEET_ENROLL=1" \
docker.elastic.co/beats/elastic-agent:8.7.0-SNAPSHOT
```
(Note: make sure to also change the version according to what you are using)

At this point you should see your agent working in your docker and under `Confirm agent enrollment` section of the Flyout.

## Discover
Go to Discover and make sure you have docs in the `findings` and `findings-latest` indecies, `findings-latest` can take a few minutes to populate.
If you do, congrats you have a working CSPM cloud deployment.
