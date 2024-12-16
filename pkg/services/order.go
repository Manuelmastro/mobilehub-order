package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"mobilehub-order/pkg/db"
	"mobilehub-order/pkg/models"
	"mobilehub-order/pkg/pb"

	cartpb "github.com/Manuelmastro/mobilehub-cart/pkg/pb"
	productpb "github.com/Manuelmastro/mobilehub-product/v3/pkg/pb"

	"google.golang.org/grpc"
)

type OrderServiceServer struct {
	pb.UnimplementedOrderServiceServer
	H db.Handler
}

func (s *OrderServiceServer) MakeOrder(ctx context.Context, req *pb.MakeOrderRequest) (*pb.MakeOrderResponse, error) {
	// Step 1: Fetch the cart details from the CartService
	cartServiceConn, err := grpc.Dial("localhost:50054", grpc.WithInsecure()) // Replace with the actual address
	if err != nil {
		return nil, errors.New("failed to connect to cart service")
	}
	defer cartServiceConn.Close()

	cartClient := cartpb.NewCartServiceClient(cartServiceConn)
	cartResp, err := cartClient.GetCart(ctx, &cartpb.GetCartRequest{UserId: req.UserId})
	if err != nil || len(cartResp.Items) == 0 {
		return nil, errors.New("cart is empty or unavailable")
	}

	// Step 2: Reduce stock for each product in the cart using ProductService
	productServiceConn, err := grpc.Dial("localhost:50052", grpc.WithInsecure()) // Replace with the actual address
	if err != nil {
		return nil, errors.New("failed to connect to product service")
	}
	defer productServiceConn.Close()

	productClient := productpb.NewProductServiceClient(productServiceConn)

	for _, cartItem := range cartResp.Items {

		productId, err := strconv.ParseInt(cartItem.ProductId, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid product ID %s: %v", cartItem.ProductId, err)
		}

		_, err = productClient.ReduceStock(ctx, &productpb.ReduceStockRequest{
			ProductId: productId,
			Quantity:  cartItem.Quantity,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to reduce stock for product %s: %v", cartItem.ProductId, err)
		}
	}

	// Step 3: Create the order in the database
	var orderItems []models.OrderItem
	for _, cartItem := range cartResp.Items {
		orderItems = append(orderItems, models.OrderItem{
			ProductID:   cartItem.ProductId,
			ProductName: cartItem.ProductName,
			Price:       float64(cartItem.Price),
			Quantity:    int(cartItem.Quantity),
			TotalPrice:  float64(cartItem.TotalPrice),
		})
	}

	order := models.Order{
		UserID:  req.UserId,
		Address: req.Address,
		Items:   orderItems,
		Total:   calculateTotal(orderItems),
		Payment: "COD",
		Status:  "Pending",
	}

	if err := s.H.DB.Create(&order).Error; err != nil {
		return nil, errors.New("failed to create order")
	}

	// Step 4: Clear the cart
	_, err = cartClient.ClearCart(ctx, &cartpb.ClearCartRequest{UserId: req.UserId})
	if err != nil {
		return nil, errors.New("order created but failed to clear the cart")
	}

	return &pb.MakeOrderResponse{
		Message: "Order placed successfully",
		OrderId: fmt.Sprintf("%d", order.ID),
	}, nil
}

// func (s *OrderServiceServer) MakeOrder(ctx context.Context, req *pb.MakeOrderRequest) (*pb.MakeOrderResponse, error) {
// 	// Step 1: Fetch the cart details from the CartService
// 	cartServiceConn, err := grpc.Dial("localhost:50054", grpc.WithInsecure()) // Replace with the actual address
// 	if err != nil {
// 		return nil, errors.New("failed to connect to cart service")
// 	}
// 	defer cartServiceConn.Close()

// 	cartClient := cartpb.NewCartServiceClient(cartServiceConn)
// 	cartResp, err := cartClient.GetCart(ctx, &cartpb.GetCartRequest{UserId: req.UserId})
// 	if err != nil || len(cartResp.Items) == 0 {
// 		return nil, errors.New("cart is empty or unavailable")
// 	}

// 	// Step 2: Create the order in the database
// 	var orderItems []models.OrderItem
// 	for _, cartItem := range cartResp.Items {
// 		orderItems = append(orderItems, models.OrderItem{
// 			ProductID:   cartItem.ProductId,
// 			ProductName: cartItem.ProductName,
// 			Price:       float64(cartItem.Price),
// 			Quantity:    int(cartItem.Quantity),
// 			TotalPrice:  float64(cartItem.TotalPrice),
// 		})
// 	}

// 	order := models.Order{
// 		UserID:  req.UserId,
// 		Address: req.Address,
// 		Items:   orderItems,
// 		Total:   calculateTotal(orderItems),
// 		Payment: "COD",
// 		Status:  "Pending",
// 	}

// 	if err := s.H.DB.Create(&order).Error; err != nil {
// 		return nil, errors.New("failed to create order")
// 	}

// 	// Step 3: Clear the cart
// 	_, err = cartClient.ClearCart(ctx, &cartpb.ClearCartRequest{UserId: req.UserId})
// 	if err != nil {
// 		return nil, errors.New("order created but failed to clear the cart")
// 	}

// 	return &pb.MakeOrderResponse{Message: "Order placed successfully"}, nil
// }

func calculateTotal(items []models.OrderItem) float64 {
	var total float64
	for _, item := range items {
		total += item.TotalPrice
	}
	return total
}

func (s *OrderServiceServer) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	var orders []models.Order
	if err := s.H.DB.Where("user_id = ?", req.UserId).Find(&orders).Error; err != nil {
		return nil, errors.New("failed to fetch orders")
	}

	var response []*pb.Order
	for _, order := range orders {
		// Fetch order items
		var orderItems []models.OrderItem
		if err := s.H.DB.Where("order_id = ?", order.ID).Find(&orderItems).Error; err != nil {
			return nil, errors.New("failed to fetch order items")
		}

		var items []*pb.Product
		for _, item := range orderItems {
			items = append(items, &pb.Product{
				ProductId:   item.ProductID,
				ProductName: item.ProductName,
				Price:       float32(item.Price),
				Quantity:    int32(item.Quantity),
			})
		}

		response = append(response, &pb.Order{
			OrderId:     fmt.Sprintf("%d", order.ID),
			UserId:      order.UserID,
			Address:     order.Address,
			TotalAmount: float32(order.Total),
			Products:    items,
			OrderStatus: order.Status,
		})
	}

	return &pb.ListOrdersResponse{Orders: response}, nil
}
