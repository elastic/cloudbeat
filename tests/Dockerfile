FROM python:3.14-slim@sha256:5b3879b6f3cb77e712644d50262d05a7c146b7312d784a18eff7ff5462e77033 as builder

WORKDIR /usr/src/app

RUN pip install poetry

RUN pip install poetry-plugin-export

COPY pyproject.toml poetry.lock ./

RUN poetry export --without-hashes -f requirements.txt --output requirements.txt


FROM python:3.14-slim@sha256:5b3879b6f3cb77e712644d50262d05a7c146b7312d784a18eff7ff5462e77033

ENV PYTHONUNBUFFERED=1

WORKDIR /usr/src/app

COPY --from=builder /usr/src/app/requirements.txt .

COPY test_environments/cloudbeat-pytest.yml .

RUN pip install -r requirements.txt

COPY . .

CMD [ "pytest", "-v", "-s", "--alluredir=./reports"]
