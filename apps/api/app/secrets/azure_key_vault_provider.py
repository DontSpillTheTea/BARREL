from .provider import SecretProvider

class AzureKeyVaultProvider(SecretProvider):
    def get_secret(self, key: str) -> str:
        # Placeholder for future Azure Key Vault integration
        raise NotImplementedError("Azure Key Vault integration not yet implemented.")
