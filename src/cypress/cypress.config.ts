import { defineConfig } from "cypress";
import * as archiver from 'archiver';

function zipFolderInMemory(folderPath: string): Promise<Buffer> {
  return new Promise((resolve, reject) => {
    const archive = archiver('zip', { zlib: { level: 9 } });
    const chunks: Uint8Array[] = [];

    archive.on('data', (chunk) => chunks.push(chunk));
    archive.on('end', () => resolve(Buffer.concat(chunks)));
    archive.on('error', (err) => reject(err));

    archive.directory(folderPath, false);
    archive.finalize();
  });
}

export default defineConfig({
  e2e: {
    setupNodeEvents(on, config) {
      on('task', {
        zipFolderInMemory(folderPath) {
          return zipFolderInMemory(folderPath);
        },
      });
    },
  },
  watchForFileChanges: false,
});

