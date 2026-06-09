try:
    from fastapi import FastAPI
    app = FastAPI(
        title="BARREL API",
        description="Beverage Alcohol Review & Regulatory Evidence Logger API",
        version="0.1.0"
    )

    @app.get("/")
    def read_root():
        return {"message": "Welcome to BARREL API (Placeholder)"}

except ImportError:
    print("FastAPI is not installed yet. Please run `make setup` when dependencies are defined.")
    # Fallback minimal app or just pass
    pass

if __name__ == "__main__":
    print("Run this using uvicorn: `uvicorn apps.api.app.main:app --reload`")
