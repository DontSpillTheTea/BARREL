import os
from .provider import SecretProvider

class EnvSecretProvider(SecretProvider):
    def get_secret(self, key: str) -> str:
        return os.environ.get(key, "")
