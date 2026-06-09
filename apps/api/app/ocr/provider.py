class OCRProvider:
    def extract_text(self, image_path: str) -> str:
        raise NotImplementedError
