package handlers

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/gin-gonic/gin"
	"github.com/nu7hatch/gouuid"
	"io"
	"jumbopetstore/models"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func AddPet(svc *dynamodb.DynamoDB) gin.HandlerFunc {

	return func(c *gin.Context) {

		var res models.ApiResponse
		var pet models.Pet

		bindErr := c.BindJSON(&pet)

		if bindErr != nil {
			fmt.Println(bindErr.Error())
			res = models.ApiResponse{http.StatusMethodNotAllowed, "error", "Invalid input"}
			c.JSON(http.StatusMethodNotAllowed, res)
			return
		}

		av, dynamodbMarshalErr := dynamodbattribute.MarshalMap(pet)
		if dynamodbMarshalErr != nil {
			fmt.Println(dynamodbMarshalErr.Error())
			res = models.ApiResponse{http.StatusInternalServerError, "error", dynamodbMarshalErr.Error()}
			c.JSON(http.StatusInternalServerError, res)
			return
		}

		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String(os.Getenv("AWS_DYNAMO_DB_TABLE")),
		}
		_, dynamodbMarshalPutErr := svc.PutItem(input)

		if dynamodbMarshalPutErr != nil {
			fmt.Println(dynamodbMarshalPutErr.Error())
			res = models.ApiResponse{http.StatusInternalServerError, "error", dynamodbMarshalPutErr.Error()}
			c.JSON(http.StatusInternalServerError, res)
			return
		}

		c.JSON(http.StatusOK, pet)
	}
}

/*we need to use this because gin does not have wild card matching yet*/
func ResolveRouteConflict(svc *dynamodb.DynamoDB) gin.HandlerFunc {

	return func(c *gin.Context) {

		if strings.HasPrefix(c.Request.RequestURI, "/api/pet/findByStatus") {
			findPetByStatus(svc, c)
		} else {
			getByID(svc, c)
		}

	}
}

func getByID(svc *dynamodb.DynamoDB, c *gin.Context) {

	var pet models.Pet

	petIdValue := c.Param("petId")

	petId, invalidIDerr := strconv.Atoi(petIdValue)

	if invalidIDerr != nil {
		fmt.Println(invalidIDerr.Error())
		res := models.ApiResponse{http.StatusBadRequest, "error", "Invalid ID supplied"}
		c.JSON(http.StatusBadRequest, res)
		return
	}

	var filter = expression.ConditionBuilder{}
	var projection = expression.ProjectionBuilder{}

	filter = expression.Name("id").Equal(expression.Value(petId))
	projection = expression.NamesList(expression.Name("id"))

	projection = projection.AddNames(expression.Name("category"))
	projection = projection.AddNames(expression.Name("name"))
	projection = projection.AddNames(expression.Name("photoUrls"))
	projection = projection.AddNames(expression.Name("tags"))
	projection = projection.AddNames(expression.Name("status"))

	expr, expressionErr := expression.NewBuilder().WithFilter(filter).WithProjection(projection).Build()

	if expressionErr != nil {
		fmt.Println(expressionErr.Error())
		res := models.ApiResponse{http.StatusInternalServerError, "error", "Expression error"}
		c.JSON(http.StatusInternalServerError, res)
		return
	}

	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(os.Getenv("AWS_DYNAMO_DB_TABLE")),
	}

	result, scanErr := svc.Scan(params)

	if scanErr != nil {
		fmt.Println(scanErr.Error())
		res := models.ApiResponse{1, "error", "Scan error"}
		c.JSON(http.StatusInternalServerError, res)
		return
	}

	if len(result.Items) < 1 {
		res := models.ApiResponse{1, "error", "Pet not found"}
		c.JSON(http.StatusNotFound, res)
		return
	}

	unmarshallingErr := dynamodbattribute.UnmarshalMap(result.Items[0], &pet)

	if unmarshallingErr != nil {
		fmt.Println(unmarshallingErr.Error())
		res := models.ApiResponse{1, "error", "Unmarshalling error"}
		c.JSON(http.StatusInternalServerError, res)
		return
	}

	c.JSON(http.StatusOK, pet)

}
func findPetByStatus(svc *dynamodb.DynamoDB, c *gin.Context) {

	var pets []models.Pet

	statusMap := c.Request.URL.Query()
	statusArray := statusMap["status"]

	if len(statusArray) < 1 {
		res := models.ApiResponse{http.StatusBadRequest, "error", "Invalid status value"}
		c.JSON(http.StatusBadRequest, res)
		return
	}

	var filter = expression.ConditionBuilder{}
	var projection = expression.ProjectionBuilder{}

	filter = expression.Name("id").NotEqual(expression.Value(""))

	projection = expression.NamesList(expression.Name("id"))

	/*TODO they might be a simpler/better may to do expressions ?*/
	/*	for _, value := range statusArray {
	 }*/
	if len(statusArray) > 0 {
		filter = filter.And(filter, expression.Name("status").Equal(expression.Value(statusArray[0])))
	}
	if len(statusArray) > 1 {
		filter = filter.Or(filter, expression.Name("status").Equal(expression.Value(statusArray[1])))
	}
	if len(statusArray) > 2 {
		filter = filter.Or(filter, expression.Name("status").Equal(expression.Value(statusArray[2])))
	}

	projection = projection.AddNames(expression.Name("category"))
	projection = projection.AddNames(expression.Name("name"))
	projection = projection.AddNames(expression.Name("photoUrls"))
	projection = projection.AddNames(expression.Name("tags"))
	projection = projection.AddNames(expression.Name("status"))

	expr, expressionErr := expression.NewBuilder().WithFilter(filter).WithProjection(projection).Build()

	if expressionErr != nil {
		fmt.Println(expressionErr.Error())
		res := models.ApiResponse{http.StatusInternalServerError, "error", "Expression error"}
		c.JSON(http.StatusInternalServerError, res)
		return
	}

	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(os.Getenv("AWS_DYNAMO_DB_TABLE")),
	}

	result, scanErr := svc.Scan(params)

	if scanErr != nil {
		fmt.Println(scanErr.Error())
		res := models.ApiResponse{http.StatusInternalServerError, "error", "Scan error"}
		c.JSON(http.StatusInternalServerError, res)
		return
	}

	if len(result.Items) < 1 {
		res := models.ApiResponse{1, "error", "Pet not found"}
		c.JSON(http.StatusNotFound, res)
		return
	}

	unmarshallingErr := dynamodbattribute.UnmarshalListOfMaps(result.Items, &pets)

	if unmarshallingErr != nil {
		fmt.Println(unmarshallingErr.Error())
		res := models.ApiResponse{http.StatusInternalServerError, "error", "Unmarshalling error"}
		c.JSON(http.StatusInternalServerError, res)
		return
	}

	c.JSON(http.StatusOK, pets)

}

func DeletePet(svc *dynamodb.DynamoDB) gin.HandlerFunc {

	return func(c *gin.Context) {

		petIdValue := c.Param("petId")

		petId, invalidIDerr := strconv.Atoi(petIdValue)

		if invalidIDerr != nil {
			fmt.Println(invalidIDerr.Error())
			res := models.ApiResponse{1, "error", "Invalid ID supplied"}
			c.JSON(http.StatusBadRequest, res)
			return
		}
		var res models.ApiResponse

		input := &dynamodb.DeleteItemInput{
			Key: map[string]*dynamodb.AttributeValue{
				"id": {
					N: aws.String(strconv.Itoa(petId)),
				},
			},
			ReturnValues: aws.String("ALL_OLD"),
			TableName:    aws.String(os.Getenv("AWS_DYNAMO_DB_TABLE")),
		}

		output, deleteErr := svc.DeleteItem(input)

		if output.Attributes == nil {
			res := models.ApiResponse{1, "error", "Pet not found"}
			c.JSON(http.StatusNotFound, res)
			return
		}

		if deleteErr != nil {
			fmt.Println(deleteErr.Error())
			res := models.ApiResponse{1, "error", ""}
			c.JSON(http.StatusInternalServerError, res)
			return
		}

		info := make(map[string]int)
		info["id"] = petId

		res = models.ApiResponse{1, "info", info}

		c.JSON(http.StatusOK, res)
	}
}

func UpdatePetFormData(svc *dynamodb.DynamoDB) gin.HandlerFunc {

	return func(c *gin.Context) {

		var pet models.Pet

		petIdValue := c.Param("petId")
		name := c.DefaultPostForm("name", "")
		status := c.DefaultPostForm("status", "")

		petId, invalidIDerr := strconv.Atoi(petIdValue)

		if invalidIDerr != nil {
			fmt.Println(invalidIDerr.Error())
			res := models.ApiResponse{http.StatusBadRequest, "error", "Invalid input"}
			c.JSON(http.StatusBadRequest, res)
			return
		}

		pet.Id = int64(petId)
		pet.Name = name
		pet.Status = status

		/*we need to use a longer expresion because name and status a reserved word in dynamoDB*/
		input := &dynamodb.UpdateItemInput{
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":name": {
					S: aws.String(name),
				},
				":status": {
					S: aws.String(status),
				},
			},
			ExpressionAttributeNames: map[string]*string{
				"#petName":   aws.String("name"),
				"#petStatus": aws.String("status"),
			},
			Key: map[string]*dynamodb.AttributeValue{
				"id": {
					N: aws.String(strconv.Itoa(petId)),
				},
			},
			ConditionExpression: aws.String("attribute_exists(id)"),
			UpdateExpression:    aws.String("set #petName = :name, #petStatus = :status"),
			TableName:           aws.String(os.Getenv("AWS_DYNAMO_DB_TABLE")),
		}

		_, dynamodbMarshalPutErr := svc.UpdateItem(input)

		if dynamodbMarshalPutErr != nil {
			res := models.ApiResponse{1, "error", "Pet not found"}
			c.JSON(http.StatusNotFound, res)
			/*		fmt.Println(dynamodbMarshalPutErr.Error())
					res = models.ApiResponse{http.StatusInternalServerError, "error", dynamodbMarshalPutErr.Error()}
					c.JSON(http.StatusInternalServerError, res)*/
			return
		}

		c.JSON(http.StatusOK, pet)
	}
}

func UpLoadImage(svc *dynamodb.DynamoDB) gin.HandlerFunc {

	return func(c *gin.Context) {

		var pet models.Pet

		petIdValue := c.Param("petId")
		file, header, invalidFileerr := c.Request.FormFile("file")

		petId, invalidIDerr := strconv.Atoi(petIdValue)

		if invalidIDerr != nil {
			fmt.Println(invalidIDerr.Error())
			res := models.ApiResponse{http.StatusBadRequest, "error", "Invalid input"}
			c.JSON(http.StatusBadRequest, res)
			return
		}

		if invalidFileerr != nil {
			fmt.Println(invalidFileerr.Error())
			res := models.ApiResponse{http.StatusBadRequest, "error", "Invalid file"}
			c.JSON(http.StatusBadRequest, res)
			return
		}

		filename := header.Filename
		fmt.Println(header.Filename)
		uid, uidErr := uuid.NewV4()
		if uidErr != nil {
			fmt.Println(uidErr.Error())
			res := models.ApiResponse{http.StatusInternalServerError, "error", "uid error"}
			c.JSON(http.StatusInternalServerError, res)
		}

		path := "static/" + petIdValue + "/" + uid.String() + "/"
		createPathErr := os.MkdirAll(path, os.ModePerm)
		if createPathErr != nil {
			fmt.Println(createPathErr.Error())
			res := models.ApiResponse{http.StatusInternalServerError, "error", "create Path error"}
			c.JSON(http.StatusInternalServerError, res)
		}

		out, writeFile := os.Create(path + filename)

		if writeFile != nil {
			fmt.Println(writeFile.Error())
			res := models.ApiResponse{http.StatusInternalServerError, "error", "unable to write file"}
			c.JSON(http.StatusInternalServerError, res)
		}

		defer out.Close()

		_, copyErr := io.Copy(out, file)
		if copyErr != nil {
			fmt.Println(copyErr.Error())
			res := models.ApiResponse{http.StatusInternalServerError, "error", "unable to write file"}
			c.JSON(http.StatusInternalServerError, res)
		}

		if invalidIDerr != nil {
			fmt.Println(invalidIDerr.Error())
			res := models.ApiResponse{http.StatusBadRequest, "error", "Invalid input"}
			c.JSON(http.StatusBadRequest, res)
			return
		}

		pet.Id = int64(petId)

		host := c.Request.Host + "/"
		imagePath := host + path + filename

		av := &dynamodb.AttributeValue{
			S: aws.String(imagePath),
		}

		var photoUrls []*dynamodb.AttributeValue
		photoUrls = append(photoUrls, av)

		input := &dynamodb.UpdateItemInput{
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":photoUrl": {
					L: photoUrls,
				},
				":empty_list": {
					L: []*dynamodb.AttributeValue{},
				},
			},
			Key: map[string]*dynamodb.AttributeValue{
				"id": {
					N: aws.String(strconv.Itoa(petId)),
				},
			},
			ConditionExpression: aws.String("attribute_exists(id)"),
			UpdateExpression:    aws.String("SET photoUrls = list_append(if_not_exists(photoUrls, :empty_list), :photoUrl)"),
			TableName:           aws.String(os.Getenv("AWS_DYNAMO_DB_TABLE")),
		}

		_, dynamodbMarshalPutErr := svc.UpdateItem(input)

		if dynamodbMarshalPutErr != nil {
			res := models.ApiResponse{1, "error", "Pet not found"}
			c.JSON(http.StatusNotFound, res)
			fmt.Println(dynamodbMarshalPutErr.Error())
			/*	res = models.ApiResponse{http.StatusInternalServerError, "error", dynamodbMarshalPutErr.Error()}
				c.JSON(http.StatusInternalServerError, res)*/
			return
		}

		info := make(map[string]string)
		info["image"] = imagePath

		res := models.ApiResponse{1, "info", info}

		c.JSON(http.StatusOK, res)
	}
}
