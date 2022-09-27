/**
 * @fileoverview gRPC-Web generated client stub for contents
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck



const grpc = {};
grpc.web = require('grpc-web');

const proto = {};
proto.contents = require('./content_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.contents.NexivilClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options.format = 'text';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.contents.NexivilPromiseClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options.format = 'text';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.contents.ContentRequest,
 *   !proto.contents.ContentResponse>}
 */
const methodDescriptor_Nexivil_NexivilContent = new grpc.web.MethodDescriptor(
  '/contents.Nexivil/NexivilContent',
  grpc.web.MethodType.SERVER_STREAMING,
  proto.contents.ContentRequest,
  proto.contents.ContentResponse,
  /**
   * @param {!proto.contents.ContentRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.contents.ContentResponse.deserializeBinary
);


/**
 * @param {!proto.contents.ContentRequest} request The request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!grpc.web.ClientReadableStream<!proto.contents.ContentResponse>}
 *     The XHR Node Readable Stream
 */
proto.contents.NexivilClient.prototype.nexivilContent =
    function(request, metadata) {
  return this.client_.serverStreaming(this.hostname_ +
      '/contents.Nexivil/NexivilContent',
      request,
      metadata || {},
      methodDescriptor_Nexivil_NexivilContent);
};


/**
 * @param {!proto.contents.ContentRequest} request The request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!grpc.web.ClientReadableStream<!proto.contents.ContentResponse>}
 *     The XHR Node Readable Stream
 */
proto.contents.NexivilPromiseClient.prototype.nexivilContent =
    function(request, metadata) {
  return this.client_.serverStreaming(this.hostname_ +
      '/contents.Nexivil/NexivilContent',
      request,
      metadata || {},
      methodDescriptor_Nexivil_NexivilContent);
};


module.exports = proto.contents;

