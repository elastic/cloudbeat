FROM python:3.15-slim@sha256:f5b92f2a669a4c347044f29367f31983280b648250ca94dea66f98c53bc87e58 as builder

WORKDIR /usr/src/app

RUN pip install poetry

RUN pip install poetry-plugin-export

COPY pyproject.toml poetry.lock ./

RUN poetry export --without-hashes -f requirements.txt --output requirements.txt


FROM python:3.15-slim@sha256:f5b92f2a669a4c347044f29367f31983280b648250ca94dea66f98c53bc87e58

ENV PYTHONUNBUFFERED=1

WORKDIR /usr/src/app

COPY --from=builder /usr/src/app/requirements.txt .

COPY test_environments/cloudbeat-pytest.yml .

RUN pip install -r requirements.txt

COPY . .

CMD [ "pytest", "-v", "-s", "--alluredir=./reports"]
