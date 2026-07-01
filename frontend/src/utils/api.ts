export const fetchWithRetry = async (url: string, options: any, retries = 3) => {
  let attempt = 0;
  while (attempt < retries) {
    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 10000); // 10s timeout
      const response = await fetch(url, { ...options, signal: controller.signal });
      clearTimeout(timeoutId);
      if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
      return response;
    } catch (error) {
      attempt++;
      if (attempt >= retries) throw error;
      await new Promise((resolve) => setTimeout(resolve, Math.pow(2, attempt) * 1000)); // Exponential backoff
    }
  }
};
