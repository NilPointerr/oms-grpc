const { ItemRequest, GetItemRequest, DeleteItemRequest } = require('./cmd/oms-api/protobufJs/oms_items_pb.js');
const { OMSServiceClient } = require('./cmd/oms-api/protobufJs/oms_items_grpc_web_pb.js'); // Replace with your service client

// Initialize gRPC client
const client = new OMSServiceClient('http://localhost:8089'); // Adjust the URL to your gRPC proxy

/**
 * Create an item
 * @param {string} name
 * @param {number} quantity
 */
function createItem(name, quantity) {
  const request = new ItemRequest();
  request.setName(name);
  request.setQuantity(quantity);

  client.createItem(request, {}, (err, response) => {
    if (err) {
      console.error('Error creating item:', err.message);
      return;
    }
    console.log('Item created successfully:', response.toObject());
  });
}

/**
 * Get an item by ID
 * @param {string} itemId
 */
function getItem(itemId) {
  const request = new GetItemRequest();
  request.setId(itemId);

  client.getItem(request, {}, (err, response) => {
    if (err) {
      console.error('Error fetching item:', err.message);
      return;
    }
    console.log('Fetched item:', response.toObject());
  });
}

/**
 * Delete an item
 * @param {string} itemId
 */
function deleteItem(itemId) {
  const request = new DeleteItemRequest();
  request.setId(itemId);

  client.deleteItem(request, {}, (err, response) => {
    if (err) {
      console.error('Error deleting item:', err.message);
      return;
    }
    console.log('Item deleted successfully:', response.getMessage());
  });
}

/**
 * Get all items
 */
function getAllItems() {
  const request = {}; // Empty request for getAllItems

  client.getAllItems(request, {}, (err, response) => {
    if (err) {
      console.error('Error fetching items:', err.message);
      return;
    }
    console.log('All items:', response.toObject());
  });
}

// Example usage
createItem('Sample Item', 10);
getItem('item123');
deleteItem('item123');
getAllItems();
