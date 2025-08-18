import express, { Request, Response } from "express";
import detect from "detect-port";

const DEFAULT_PORT = parseInt(process.env.PORT || "3000", 10);
const app = express();

async function startServer() {
  const port = await detect(DEFAULT_PORT);

  if (port !== DEFAULT_PORT) {
    console.error(`❌ Port ${DEFAULT_PORT} is in use. Try running on port ${port}`);
    process.exit(1);
  }

  app.get("/", (_req: Request, res: Response) => {
    res.send("Hello World from Express + GitHub + TS!");
  });

  app.listen(port, () => {
    console.log(`🚀 Server running on http://localhost:${port}`);
  });
}

startServer().catch((err) => {
  console.error("❌ Failed to start server:", err);
  process.exit(1);
});

