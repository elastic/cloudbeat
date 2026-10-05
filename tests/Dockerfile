FROM python:3.14-slim@sha256:ce40764625a4ff50df3548277632e7f96c4e77fe75fa848aae9885476e7df5a4 as builder

WORKDIR /usr/src/app

RUN pip install poetry

RUN pip install poetry-plugin-export

COPY pyproject.toml poetry.lock ./

RUN poetry export --without-hashes -f requirements.txt --output requirements.txt


FROM python:3.14-slim@sha256:ce40764625a4ff50df3548277632e7f96c4e77fe75fa848aae9885476e7df5a4

ENV PYTHONUNBUFFERED=1

WORKDIR /usr/src/app

COPY --from=builder /usr/src/app/requirements.txt .

COPY test_environments/cloudbeat-pytest.yml .

RUN pip install -r requirements.txt

COPY . .

CMD [ "pytest", "-v", "-s", "--alluredir=./reports"]
