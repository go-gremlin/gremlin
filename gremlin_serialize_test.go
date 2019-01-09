package gremlin

// import (
// 	"testing"
// )

// func TestSerializeVertexes(t *testing.T) {

// 	inputString := `[{"@type":"g:Vertex","@value":{"id":"https://sqs.us-west-2.amazonaws.com/496584544324/sqs-cbi_prd-dirty-investor-profiles-queue","label":"queue","properties":{"queue_name":[{"@type":"g:VertexProperty","@value":{"id":{"@type":"g:Int32","@value":247643139},"value":"sqs-cbi_prd-dirty-investor-profiles","label":"queue_name"}}],"dead_letter_target_arn":[{"@type":"g:VertexProperty","@value":{"id":{"@type":"g:Int32","@value":-929173958},"value":"arn:aws:sqs:us-west-2:496584544324:sqs-cbi_prd-dirty-investor-profiles-dead","label":"dead_letter_target_arn"}}],"queue_url":[{"@type":"g:VertexProperty","@value":{"id":{"@type":"g:Int32","@value":-230068604},"value":"https://sqs.us-west-2.amazonaws.com/496584544324/sqs-cbi_prd-dirty-investor-profiles","label":"queue_url"}}],"health":[{"@type":"g:VertexProperty","@value":{"id":{"@type":"g:Int32","@value":-790749087},"value":"1","label":"health"}}],"oldest_message_seconds":[{"@type":"g:VertexProperty","@value":{"id":{"@type":"g:Int32","@value":-383438945},"value":"0.000000","label":"oldest_message_seconds"}}],"create_date":[{"@type":"g:VertexProperty","@value":{"id":{"@type":"g:Int32","@value":1640500219},"value":"2018-12-14T21:51:20Z","label":"create_date"}}],"last_updated_date":[{"@type":"g:VertexProperty","@value":{"id":{"@type":"g:Int32","@value":-195565151},"value":"2019-01-08T22:33:25Z","label":"last_updated_date"}}],"dead_letter_queue_health":[{"@type":"g:VertexProperty","@value":{"id":{"@type":"g:Int32","@value":1477226235},"value":"GREEN","label":"dead_letter_queue_health"}}],"oldest_message_health":[{"@type":"g:VertexProperty","@value":{"id":{"@type":"g:Int32","@value":-482108672},"value":"UNKNOWN","label":"oldest_message_health"}}],"dead_letter_queue_messages":[{"@type":"g:VertexProperty","@value":{"id":{"@type":"g:Int32","@value":-384491714},"value":"0","label":"dead_letter_queue_messages"}}]}}}]`
// 	inputString = `[{"@type":"g:Edge","@value":{"id":"f8b3d759-fec1-ac41-0588-f7e5590d1f86","label":"moves_data","inVLabel":"topic","outVLabel":"queue","inV":"arn:aws:sns:us-west-2:496584544324:company-deletes-prd-topic","outV":"https://sqs.us-west-2.amazonaws.com/496584544324/searchconsumer-cs-company-deletes-prd-queue"}}]`
// 	// TODO: empty strings for property values will cause invalid json
// 	// make so it can handle that case

// 	result, _ := SerializeVertexes(inputString)
// 	result2, _ := SerializeEdges(inputString)
// 	result3, _ := SerializeResponse(inputString)
// 	t.Error("result string", result, "result2", result2, "result3", result3)
// }
