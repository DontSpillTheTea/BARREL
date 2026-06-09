class SecretProvider:
    def get_secret(self, key: str) -> str:
        raise NotImplementedError
