FROM python:3.12-slim

WORKDIR /code

# Instalar dependencias del sistema para matplotlib
RUN apt-get update && apt-get install -y \
    libgl1 \
    libglib2.0-0 \
    && rm -rf /var/lib/apt/lists/*

COPY ./requirements.txt /code/requirements.txt
RUN pip install --no-cache-dir --upgrade -r /code/requirements.txt

COPY . /code/

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]