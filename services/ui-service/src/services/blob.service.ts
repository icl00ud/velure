class BlobService {
  async getBase64FromUrl(url: string): Promise<string> {
    try {
      const response = await fetch(url);
      if (!response.ok) {
        throw new Error("Erro ao buscar blob");
      }

      const blob = await response.blob();

      return new Promise((resolve, reject) => {
        const reader = new FileReader();

        reader.onload = () => {
          const base64String = reader.result as string;
          resolve(base64String.split(",")[1]);
        };

        reader.onerror = (error) => {
          reject(error);
        };

        reader.readAsDataURL(blob);
      });
    } catch (error) {
      console.error("Erro ao converter URL para base64:", error);
      throw error;
    }
  }

  async downloadBlob(url: string, filename?: string): Promise<void> {
    try {
      const response = await fetch(url);
      if (!response.ok) {
        throw new Error("Erro ao baixar arquivo");
      }

      const blob = await response.blob();
      const downloadUrl = URL.createObjectURL(blob);

      const link = document.createElement("a");
      link.href = downloadUrl;
      link.download = filename || "download";
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);

      URL.revokeObjectURL(downloadUrl);
    } catch (error) {
      console.error("Erro ao baixar arquivo:", error);
      throw error;
    }
  }
}

export const blobService = new BlobService();
