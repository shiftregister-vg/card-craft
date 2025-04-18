import { Form, useActionData, useNavigation } from "@remix-run/react";
import { ActionFunctionArgs, json, unstable_parseMultipartFormData, UploadHandlerPart } from "@remix-run/node";
import { requireUser } from "~/utils/auth.server.js";

interface ImportResponse {
  success?: boolean;
  error?: string;
  result?: {
    totalCards: number;
    importedCards: number;
    updatedCards: number;
    errors: string[];
  };
}

const GRAPHQL_ENDPOINT = 'http://localhost:8080/query';

const IMPORT_MUTATION = `
  mutation ImportCollection($input: ImportSource!, $file: Upload!) {
    importCollection(input: $input, file: $file) {
      totalCards
      importedCards
      updatedCards
      errors
    }
  }
`;

export async function action({ request }: ActionFunctionArgs) {
  await requireUser(request);

  try {
    const formData = await unstable_parseMultipartFormData(request, async ({ name, filename, data }: UploadHandlerPart) => {
      if (name !== "file") {
        return undefined;
      }

      // Check if file is CSV
      if (!filename?.toLowerCase().endsWith('.csv')) {
        throw new Error('Please upload a CSV file');
      }

      // Convert stream to text
      const chunks = [];
      for await (const chunk of data) {
        chunks.push(chunk);
      }
      const buffer = Buffer.concat(chunks);
      return buffer.toString();
    });

    const fileContent = formData.get("file");
    if (!fileContent) {
      return json<ImportResponse>({ error: "No file uploaded" }, { status: 400 });
    }

    // Get token from cookie
    const cookie = request.headers.get('Cookie');
    const token = cookie
      ?.split('; ')
      .find(row => row.startsWith('token='))
      ?.split('=')[1];

    if (!token) {
      return json<ImportResponse>({ error: "Not authenticated" }, { status: 401 });
    }

    // Make GraphQL mutation call with multipart form data
    const form = new FormData();
    form.append('operations', JSON.stringify({
      query: IMPORT_MUTATION,
      variables: {
        input: {
          source: "tcgcollector",
          format: "csv"
        },
        file: null
      }
    }));
    form.append('map', JSON.stringify({ "0": ["variables.file"] }));
    form.append('0', new Blob([fileContent as string], { type: 'text/csv' }), 'collection.csv');

    const response = await fetch(GRAPHQL_ENDPOINT, {
      method: "POST",
      headers: {
        "Authorization": `Bearer ${token}`,
      },
      credentials: 'include',
      body: form
    });

    if (!response.ok) {
      const text = await response.text();
      throw new Error(`GraphQL request failed: ${text}`);
    }

    const { data, errors } = await response.json();

    if (errors) {
      return json<ImportResponse>({ 
        error: errors[0].message 
      }, { status: 400 });
    }

    return json<ImportResponse>({ 
      success: true,
      result: data.importCollection
    });

  } catch (error) {
    return json<ImportResponse>({ 
      error: error instanceof Error ? error.message : "Failed to process file" 
    }, { status: 400 });
  }
}

export default function Import() {
  const actionData = useActionData<ImportResponse>();
  const navigation = useNavigation();
  const isSubmitting = navigation.state === "submitting";

  return (
    <div className="max-w-2xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-6">Import Collection</h1>
      
      <Form method="post" encType="multipart/form-data" className="space-y-4">
        <div>
          <label htmlFor="file" className="block text-sm font-medium text-gray-700 mb-2">
            Choose CSV File
          </label>
          <input
            type="file"
            id="file"
            name="file"
            accept=".csv"
            className="block w-full text-sm text-gray-500
              file:mr-4 file:py-2 file:px-4
              file:rounded-md file:border-0
              file:text-sm file:font-semibold
              file:bg-blue-50 file:text-blue-700
              hover:file:bg-blue-100"
            required
          />
          <p className="mt-2 text-sm text-gray-500">
            Currently supports TCGCollector export format
          </p>
        </div>

        {actionData?.error && (
          <div className="text-red-600 text-sm">{actionData.error}</div>
        )}

        {actionData?.result && (
          <div className="bg-green-50 p-4 rounded-md">
            <h3 className="text-sm font-medium text-green-800">Import Successful</h3>
            <ul className="mt-2 text-sm text-green-700">
              <li>Total cards processed: {actionData.result.totalCards}</li>
              <li>New cards imported: {actionData.result.importedCards}</li>
              <li>Existing cards updated: {actionData.result.updatedCards}</li>
            </ul>
            {actionData.result.errors.length > 0 && (
              <div className="mt-2">
                <h4 className="text-sm font-medium text-green-800">Warnings:</h4>
                <ul className="mt-1 text-sm text-green-700 list-disc list-inside">
                  {actionData.result.errors.map((error, index) => (
                    <li key={index}>{error}</li>
                  ))}
                </ul>
              </div>
            )}
          </div>
        )}

        <button
          type="submit"
          disabled={isSubmitting}
          className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 
            disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isSubmitting ? "Importing..." : "Import"}
        </button>
      </Form>
    </div>
  );
} 