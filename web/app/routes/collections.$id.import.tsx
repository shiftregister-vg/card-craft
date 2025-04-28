import { json, LoaderFunctionArgs, ActionFunctionArgs } from '@remix-run/node';
import { Form, useActionData, useLoaderData, useNavigation, useParams } from '@remix-run/react';
import { createServerClient } from '../lib/urql.js';
import { requireUser } from '../utils/auth.server.js';
import { COLLECTION_QUERY } from '../graphql/collections.js';
import { gql } from '@urql/core';

const BULK_IMPORT_MUTATION = gql`
  mutation BulkImportCardsToCollection($collectionId: ID!, $file: Upload!) {
    bulkImportCardsToCollection(collectionId: $collectionId, file: $file) {
      success
      importedCount
      errors {
        cardId
        message
      }
    }
  }
`;

const FIND_CARD = gql`
  query FindCard($game: String!, $setCode: String!, $number: String!) {
    searchCards(
      game: $game
      setCode: $setCode
      name: $number
      pageSize: 100
    ) {
      cards {
        id
        name
        setCode
        number
      }
      totalCount
      page
      pageSize
    }
  }
`;

type ActionData = 
  | { success: true; count: number; errors?: Array<{ cardId: string; message: string }> }
  | { error: string };

export async function loader({ request, params }: LoaderFunctionArgs) {
  await requireUser(request);
  const serverClient = createServerClient(request);
  
  if (!params.id) {
    throw new Response('Collection ID is required', { status: 400 });
  }

  // Fetch collection details to verify it exists and user has access
  const { data, error } = await serverClient.query(COLLECTION_QUERY, {
    id: params.id,
  }).toPromise();
  
  if (error) {
    throw new Response('Error fetching collection', { status: 500 });
  }
  
  if (!data?.collection) {
    throw new Response('Collection not found', { status: 404 });
  }

  return json({ collection: data.collection });
}

export async function action({ request, params }: ActionFunctionArgs) {
  await requireUser(request);
  const serverClient = createServerClient(request);
  const formData = await request.formData();
  const file = formData.get('file') as File;
  const collectionId = params.id;

  if (!file) {
    return json<ActionData>({ error: 'No file uploaded' }, { status: 400 });
  }

  if (!collectionId) {
    return json<ActionData>({ error: 'No collection ID provided' }, { status: 400 });
  }

  try {
    // Send the raw file to the server for processing
    const { data: importData, error: importError } = await serverClient.mutation(BULK_IMPORT_MUTATION, {
      collectionId,
      file,
    }).toPromise();

    if (importError) {
      throw new Error('Failed to import cards');
    }

    return json<ActionData>({
      success: true,
      count: importData.bulkImportCardsToCollection.importedCount,
      errors: importData.bulkImportCardsToCollection.errors,
    });
  } catch (error) {
    console.error('Error processing CSV:', error);
    return json<ActionData>({ error: 'Error processing CSV file' }, { status: 500 });
  }
}

export default function ImportCollection() {
  const { id } = useParams();
  const actionData = useActionData<ActionData>();
  const navigation = useNavigation();
  const isSubmitting = navigation.state === 'submitting';

  return (
    <div className="min-h-screen bg-gray-100">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="bg-white shadow sm:rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h3 className="text-lg leading-6 font-medium text-gray-900">
              Import Cards from TCG Collector
            </h3>
            <div className="mt-2 max-w-xl text-sm text-gray-500">
              <p>
                Upload a CSV export from TCG Collector to import your cards into this collection.
                The CSV should include the following columns:
              </p>
              <ul className="list-disc pl-5 mt-2">
                <li>Card Name</li>
                <li>Set Code</li>
                <li>Card Number</li>
                <li>Quantity</li>
                <li>Condition</li>
                <li>Foil</li>
                <li>Notes</li>
              </ul>
            </div>
            <div className="mt-5">
              <Form method="post" encType="multipart/form-data">
                <div className="space-y-4">
                  <div>
                    <label htmlFor="file" className="block text-sm font-medium text-gray-700">
                      CSV File
                    </label>
                    <div className="mt-1">
                      <input
                        type="file"
                        name="file"
                        id="file"
                        accept=".csv"
                        required
                        className="block w-full text-sm text-gray-500
                          file:mr-4 file:py-2 file:px-4
                          file:rounded-md file:border-0
                          file:text-sm file:font-semibold
                          file:bg-blue-50 file:text-blue-700
                          hover:file:bg-blue-100"
                      />
                    </div>
                  </div>
                  <div>
                    <button
                      type="submit"
                      disabled={isSubmitting}
                      className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
                    >
                      {isSubmitting ? 'Importing...' : 'Import Cards'}
                    </button>
                  </div>
                </div>
              </Form>
            </div>
            {actionData && (
              <div className="mt-4">
                {'success' in actionData ? (
                  <div className="rounded-md bg-green-50 p-4">
                    <div className="flex">
                      <div className="flex-shrink-0">
                        <svg className="h-5 w-5 text-green-400" viewBox="0 0 20 20" fill="currentColor">
                          <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                        </svg>
                      </div>
                      <div className="ml-3">
                        <h3 className="text-sm font-medium text-green-800">
                          Successfully imported {actionData.count} cards
                        </h3>
                        {actionData.errors && actionData.errors.length > 0 && (
                          <div className="mt-2 text-sm text-green-700">
                            <p>Some cards could not be imported:</p>
                            <ul className="list-disc pl-5 mt-1">
                              {actionData.errors.map((error, index) => (
                                <li key={index}>
                                  {error.cardId}: {error.message}
                                </li>
                              ))}
                            </ul>
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                ) : (
                  <div className="rounded-md bg-red-50 p-4">
                    <div className="flex">
                      <div className="flex-shrink-0">
                        <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                          <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                        </svg>
                      </div>
                      <div className="ml-3">
                        <h3 className="text-sm font-medium text-red-800">
                          Error importing cards
                        </h3>
                        <div className="mt-2 text-sm text-red-700">
                          <p>{actionData.error}</p>
                        </div>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
} 