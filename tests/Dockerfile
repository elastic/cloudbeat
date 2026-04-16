FROM python:3.14-slim@sha256:bc389f7dfcb21413e72a28f491985326994795e34d2b86c8ae2f417b4e7818aa as builder

WORKDIR /usr/src/app

RUN pip install poetry

RUN pip install poetry-plugin-export

COPY pyproject.toml poetry.lock ./

RUN poetry export --without-hashes -f requirements.txt --output requirements.txt


FROM python:3.14-slim@sha256:bc389f7dfcb21413e72a28f491985326994795e34d2b86c8ae2f417b4e7818aa

ENV PYTHONUNBUFFERED=1

WORKDIR /usr/src/app

COPY --from=builder /usr/src/app/requirements.txt .

COPY test_environments/cloudbeat-pytest.yml .

RUN pip install -r requirements.txt

COPY . .

CMD [ "pytest", "-v", "-s", "--alluredir=./reports"]
