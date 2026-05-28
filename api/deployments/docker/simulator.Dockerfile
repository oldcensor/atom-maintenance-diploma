FROM python:3.12-slim
WORKDIR /app
COPY tools/simulator/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY tools/simulator/ .
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8090"]
