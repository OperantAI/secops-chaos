FROM python:3.11-slim

LABEL org.opencontainers.image.source=https://github.com/OperantAI/secops-chaos
LABEL org.opencontainers.image.description="AI Security focused Chaos Experiments"

# Install system packages needed for building
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

RUN useradd -m -u 1000 user
USER user
ENV HOME=/home/user \
    PATH=/home/user/.local/bin:$PATH

# ensures that the python output is sent straight to terminal (e.g. your container log)
ENV PYTHONUNBUFFERED 1

# https://docs.python.org/3/using/cmdline.html#envvar-PYTHONDONTWRITEBYTECODE
ENV PYTHONDONTWRITEBYTECODE 1

WORKDIR $HOME/app

COPY --chown=user:user ./cmd/secops-chaos-ai/pyproject.toml ./
COPY --chown=user:user ./cmd/secops-chaos-ai/app ./app


RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir .

COPY --chown=user:user ./cmd/secops-chaos-ai/entrypoint.sh ./

RUN chmod +x /home/user/app/entrypoint.sh

EXPOSE 8000

CMD ["/home/user/app/entrypoint.sh"]
