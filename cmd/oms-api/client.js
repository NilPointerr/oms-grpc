// Polyfill XMLHttpRequest for Node.js
const XMLHttpRequest = require('xhr2');
global.XMLHttpRequest = XMLHttpRequest; // Set it as global.XMLHttpRequest

const { ItemRequest, GetItemRequest, DeleteItemRequest } = require('./protobufJs/oms_items_pb.js');
const { omsItemServiceClient } = require('./protobufJs/oms_items_grpc_web_pb.js'); // Correct import for the client

// Initialize gRPC client
const client = new omsItemServiceClient('http://localhost:8080'); // Adjust the URL to your gRPC proxy

/**
 * Create an item
 * @param {string} name
 * @param {string} description
 * @param {number} price
 */
// function createItem(name, description, price) {
//   const request = new ItemRequest();
//   request.setName(name);  // Set name
//   request.setDescription(description);  // Set description
//   request.setPrice(price);  // Set price

//   client.createItem(request, {}, (err, response) => {
//     if (err) {
//       console.error('Error creating item:', err.message);
//       return;
//     }
//     console.log('Item created successfully:', response.toObject());
//   });
// }

/**
 * Get an item by ID
 * @param {number} itemId
 */
function getItem(itemId) {
  if (typeof itemId !== 'number') {
    console.error('Item ID must be a number');
    return;
  }

  const request = new GetItemRequest();
  request.setId(itemId);  // Ensure that 'itemId' is passed as a number, not a string

  client.getItemById(request, {}, (err, response) => {
    if (err) {
      console.error('Error fetching item:', err.message);
      return;
    }
    console.log('Fetched item:', response.toObject());
  });
}

// /**
//  * Delete an item by ID
//  * @param {number} itemId
//  */
// function deleteItem(itemId) {
//   if (typeof itemId !== 'number') {
//     console.error('Item ID must be a number');
//     return;
//   }

//   const request = new DeleteItemRequest();
//   request.setId(itemId);  // Ensure that 'itemId' is passed as a number, not a string

//   client.deleteItemById(request, {}, (err, response) => {
//     if (err) {
//       console.error('Error deleting item:', err.message);
//       return;
//     }
//     console.log('Item deleted successfully:', response.getMessage());  // Adjust based on response type
//   });
// }

// /**
//  * Get all items
//  */
// function getAllItems() {
//   const request = {}; // Empty request for getAllItems (if no parameters are needed)

//   client.getAllItems(request, {}, (err, response) => {
//     if (err) {
//       console.error('Error fetching items:', err.message);
//       return;
//     }
//     console.log('All items:', response.toObject());
//   });
// }

// Example usage
// createItem('Sample Item', 'This is a description of the sample item', 100);
getItem(1);  // Ensure that 'itemId' is a number
// deleteItem(123);  // Ensure that 'itemId' is a number
// getAllItems();
