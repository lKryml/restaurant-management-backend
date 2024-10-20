package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"path/filepath"
	"restaurant-management-backend/internal/helpers"
	"restaurant-management-backend/internal/logger"
	middleware2 "restaurant-management-backend/internal/middleware"
	"restaurant-management-backend/internal/service"
	"restaurant-management-backend/internal/types"
	"strings"
)

func (s *Server) RegisterRoutes() http.Handler {

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RedirectSlashes)
	r.Use(middleware2.JWTMiddleware(s.db.GetDB()))

	r.Route("/api/v1", func(r chi.Router) {

		r.Route("/auth", func(r chi.Router) {
			r.Post("/signup", s.SignUpHandler)
			r.Post("/login", s.LoginHandler)
		})

		r.Route("/users", func(user chi.Router) {
			user.Use(middleware2.RoleMiddleware(1))

			user.Get("/", s.indexUsersHandler)
			user.Post("/", s.createUserHandler)
			user.Get("/{id}", s.getUserHandler)
			user.Put("/{id}", s.updateUserHandler)
			user.Delete("/{id}", s.deleteUserHandler)
		})

		r.Route("/roles", func(r chi.Router) {

			r.Get("/", s.indexRolesHandler)
			r.Post("/{id}", s.grantRoleHandler)
			r.Get("/{id}", s.getRoleHandler)
			r.Delete("/{id}", s.revokeRoleHandler)
		})

		r.Route("/tables", func(r chi.Router) {
			r.Get("/", s.IndexTablesHandler)
			r.Post("/", s.AddTableHandler)
			r.Get("/{id}", s.GetTableHandler)
			r.Put("/{id}", s.UpdateTableHandler)
			r.Delete("/{id}", s.DeleteTableHandler)
		})

		r.Route("/orders", func(r chi.Router) {
			r.Get("/", s.IndexOrdersHandler)
			r.Get("/{id}", s.GetOrderHandler)
			r.Put("/{id}", s.UpdateOrderHandler)
		})

		r.Route("/items", func(r chi.Router) {
			r.Get("/", s.ListItemsHandler)
			r.Post("/", s.CreateItemHandler)
			r.Get("/{id}", s.GetItemHandler)
			r.Put("/{id}", s.UpdateItemHandler)
			r.Delete("/{id}", s.DeleteItemHandler)
		})

		r.Route("/cart", func(r chi.Router) {
			r.Get("/", s.IndexCartHandler)
			r.Post("/", s.CreateCartHandler)
			r.Delete("/", s.EmptyCartHandler)
			r.Post("/checkout", s.CheckoutHandler)
		})

		r.Route("/vendors", func(r chi.Router) {
			r.Get("/", s.IndexVendorsHandler)
			r.Post("/", s.CreateVendorHandler)
			r.Get("/{id}", s.GetVendorHandler)
			r.Put("/{id}", s.UpdateVendorHandler)
			r.Delete("/{id}", s.DeleteVendorHandler)
			r.Get("/", s.IndexVendorAdminsHandler)
			r.Post("/admin/grant", s.GrantAdminHandler)
			r.Post("/admin/revoke", s.RevokeAdminHandler)
		})

	})

	return r
}

func (s *Server) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	user := types.User{
		ID:       uuid.New(),
		Name:     r.FormValue("name"),
		Phone:    r.FormValue("phone"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	if user.Password == "" {
		helpers.HandleError(w, http.StatusBadRequest, "Password is required")
		return
	}

	imagePath, err := helpers.HandleFileUpload(r, "users")
	if err != nil {
		helpers.HandleError(w, http.StatusBadRequest, fmt.Sprintf("Error handling file upload: %v", err))
		return
	}
	user.Img = imagePath

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, "Error hashing password")
		return
	}
	user.Password = string(hashedPassword)

	createdUser, err := s.db.CreateUser(user)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to create user")
		helpers.HandleError(w, http.StatusInternalServerError, "Error creating user")
		return
	}

	err = s.db.GrantDefaultRole(createdUser.ID)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to grant default role")
		helpers.HandleError(w, http.StatusInternalServerError, "Error granting role")
		return
	}

	helpers.WriteJSONResponse(w, http.StatusCreated, createdUser)
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	if email == "" || password == "" {
		helpers.HandleError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	user, err := s.db.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helpers.HandleError(w, http.StatusUnauthorized, "Invalid email or password")
		} else {
			logger.Log.WithError(err).Error("Failed to fetch user")
			helpers.HandleError(w, http.StatusInternalServerError, "Error during login")
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		helpers.HandleError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	token, err := helpers.GenerateJWT(user.ID)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to generate token")
		helpers.HandleError(w, http.StatusInternalServerError, "Error generating token")
		return
	}

	helpers.WriteJSONResponse(w, http.StatusOK, map[string]string{"token": fmt.Sprintf("%s", token)})
}

// ///////////
func (s *Server) indexUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := s.db.ListUsers()
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}
	helpers.WriteJSONResponse(w, 200, users)
}

func (s *Server) getUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		helpers.HandleError(w, http.StatusBadRequest, "id is required")
		return
	}
	user, err := s.db.GetUserByID(id)
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.WriteJSONResponse(w, 200, user)

}

func (s *Server) createUserHandler(w http.ResponseWriter, r *http.Request) {

	user, err := service.SignUpHandler(r)
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	user, err = s.db.CreateUser(*user)
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make(map[string]interface{})
	resp["message"] = "User successfully signed up"
	resp["user"] = user
	helpers.WriteJSONResponse(w, http.StatusCreated, resp)

}

func (s *Server) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		helpers.HandleError(w, http.StatusBadRequest, "id is required")
		return
	}

	stuff2update, err := service.UserValidator(r)
	if err != nil {
		helpers.HandleError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := s.db.UpdateUser(*stuff2update, id)
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make(map[string]interface{})
	resp["message"] = "Updated user successfully!"
	resp["userID"] = id
	resp["user"] = user
	helpers.WriteJSONResponse(w, http.StatusCreated, resp)

}

func (s *Server) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := s.db.DeleteUser(id)
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make(map[string]string)
	resp["message"] = "Deleted user successfully!"
	resp["userID"] = id
	helpers.WriteJSONResponse(w, http.StatusCreated, resp)

}

/////////////////////////////

func (s *Server) serveFileHandler(w http.ResponseWriter, r *http.Request) {
	filePath := strings.TrimPrefix(r.URL.Path, "/uploads/")
	fullPath := filepath.Join("./uploads", filePath)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		helpers.HandleError(w, http.StatusNotFound, "File not found")
		return
	}
	http.ServeFile(w, r, fullPath)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJSONResponse(w, 200, s.db.Health())
}

//////////////////////////////

func (s *Server) indexRolesHandler(w http.ResponseWriter, r *http.Request) {

	roles, meta, err := s.db.FetchRoles(r.URL.Query())
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}
	helpers.WriteJSONResponse(w, http.StatusOK, types.Response{Meta: meta, Data: roles})
}

func (s *Server) getRoleHandler(w http.ResponseWriter, r *http.Request) {
	role, err := s.db.FetchRole(r.PathValue("id"))
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}
	helpers.WriteJSONResponse(w, http.StatusOK, role)
}

func (s *Server) grantRoleHandler(w http.ResponseWriter, r *http.Request) {
	userID, roleID := r.FormValue("user_id"), r.FormValue("role_id")
	if userID == "" || roleID == "" {
		helpers.HandleError(w, http.StatusBadRequest, "Missing required parameters")
		return
	}

	if err := s.db.VerifyRoleExists(roleID); err != nil {
		helpers.HandleError(w, http.StatusNotFound, "Role not found")
		return
	}

	if err := s.db.GrantRole(userID, roleID); err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, "Role already granted")
		return
	}

	helpers.WriteJSONResponse(w, http.StatusOK, "Role granted successfully")
}

func (s *Server) revokeRoleHandler(w http.ResponseWriter, r *http.Request) {
	userID, roleID := r.FormValue("user_id"), r.FormValue("role_id")
	if userID == "" || roleID == "" {
		helpers.HandleError(w, http.StatusBadRequest, "Missing required parameters")
		return
	}

	affected, err := s.db.RevokeRole(userID, roleID)
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if affected == 0 {
		helpers.HandleError(w, http.StatusNotFound, "Role not granted for user")
		return
	}

	helpers.WriteJSONResponse(w, http.StatusOK, "Role revoked successfully")
}

///////////////////////

func (s *Server) IndexOrdersHandler(w http.ResponseWriter, r *http.Request) {
	orders, meta, err := s.db.FetchOrders(r.URL.Query())
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.db.EnrichOrdersWithItems(orders)

	helpers.WriteJSONResponse(w, http.StatusOK, types.Response{Meta: meta, Data: orders})
}

func (s *Server) GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	order, err := s.db.FetchOrder(id)
	if err != nil {
		helpers.HandleError(w, http.StatusNotFound, "Order not found")
		return
	}

	s.db.AttachOrderItems(&order)

	helpers.WriteJSONResponse(w, http.StatusOK, order)
}

func (s *Server) UpdateOrderHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	status := r.FormValue("status")

	if err := s.db.UpdateOrderStatus(id, status); err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Order status updated successfully"})
}

///////////////////

func (s *Server) IndexTablesHandler(w http.ResponseWriter, r *http.Request) {
	tables, meta, err := s.db.FetchTables(r.URL.Query())
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}
	helpers.WriteJSONResponse(w, http.StatusOK, types.Response{Meta: meta, Data: tables})
}

func (s *Server) GetTableHandler(w http.ResponseWriter, r *http.Request) {
	table, err := s.db.GetTableByID(r.PathValue("id"))
	if err != nil {
		helpers.HandleError(w, http.StatusNotFound, "Table not found")
		return
	}
	helpers.WriteJSONResponse(w, http.StatusOK, table)
}

func (s *Server) AddTableHandler(w http.ResponseWriter, r *http.Request) {
	table, err := s.db.ParseTableFromForm(r)
	if err != nil {
		helpers.HandleError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.db.InsertTable(&table); err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.WriteJSONResponse(w, http.StatusCreated, table)
}

func (s *Server) UpdateTableHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	existingTable, err := s.db.GetTableByID(id)
	if err != nil {
		helpers.HandleError(w, http.StatusNotFound, "Table not found")
		return
	}

	if err := s.db.UpdateTableFromForm(&existingTable, r); err != nil {
		helpers.HandleError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.db.UpdateTable(&existingTable); err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.WriteJSONResponse(w, http.StatusOK, existingTable)
}

func (s *Server) DeleteTableHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.db.DeleteTable(id); err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}
	helpers.WriteJSONResponse(w, http.StatusNoContent, nil)
}

///////////

func (s *Server) ListItemsHandler(w http.ResponseWriter, r *http.Request) {
	items, meta, err := s.db.ListItems(r.URL.Query())
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}
	helpers.WriteJSONResponse(w, http.StatusOK, types.Response{Meta: meta, Data: items})
}

func (s *Server) CreateItemHandler(w http.ResponseWriter, r *http.Request) {
	var item types.Item
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		helpers.HandleError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	createdItem, err := s.db.CreateItem(item, r)
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.WriteJSONResponse(w, http.StatusCreated, createdItem)
}

func (s *Server) GetItemHandler(w http.ResponseWriter, r *http.Request) {
	item, err := s.db.GetItemByID(r.PathValue("id"))
	if err != nil {
		helpers.HandleError(w, http.StatusNotFound, "Item not found")
		return
	}
	helpers.WriteJSONResponse(w, http.StatusOK, item)
}

func (s *Server) DeleteItemHandler(w http.ResponseWriter, r *http.Request) {
	err := s.db.DeleteItem(r.PathValue("id"))
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}
	helpers.WriteJSONResponse(w, http.StatusOK, "Item deleted successfully")
}

func (s *Server) UpdateItemHandler(w http.ResponseWriter, r *http.Request) {
	var updates map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		helpers.HandleError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updatedItem, err := s.db.UpdateItem(r.PathValue("id"), updates, r)
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.WriteJSONResponse(w, http.StatusOK, updatedItem)
}

///////////

func (s *Server) IndexCartHandler(w http.ResponseWriter, r *http.Request) {
	userID := s.db.GetUserID(r)
	cart, err := s.db.GetCart(userID)
	if err != nil {
		helpers.HandleError(w, http.StatusNotFound, "Cart not found")
		return
	}

	cartItems, err := s.db.GetCartItems(cart.ID)
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, "Failed to fetch cart items")
		return
	}

	cart.CartItem = cartItems
	helpers.WriteJSONResponse(w, http.StatusOK, cart)
}

func (s *Server) CreateCartHandler(w http.ResponseWriter, r *http.Request) {
	userID := s.db.GetUserID(r)
	itemID, quantity, err := s.db.ParseAddCartParams(r)
	if err != nil {
		helpers.HandleError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := s.db.GetCartItem(itemID)
	if err != nil {
		helpers.HandleError(w, http.StatusNotFound, "Item does not exist")
		return
	}

	cart, err := s.db.GetOrCreateCart(userID, item.VendorId)
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, "Failed to process cart")
		return
	}

	if err := s.db.UpdateCartItem(cart.ID, itemID, quantity); err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, "Failed to update cart item")
		return
	}

	if err := s.db.RecalculateCart(cart.ID); err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, "Failed to recalculate cart")
		return
	}

	updatedCart, err := s.db.GetCart(userID)
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, "Failed to fetch updated cart")
		return
	}

	cartItems, err := s.db.GetCartItems(updatedCart.ID)
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, "Failed to fetch cart items")
		return
	}

	updatedCart.CartItem = cartItems
	helpers.WriteJSONResponse(w, http.StatusOK, updatedCart)
}

func (s *Server) EmptyCartHandler(w http.ResponseWriter, r *http.Request) {
	userID := s.db.GetUserID(r)
	if err := s.db.EmptyCart(userID); err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, "Failed to empty cart")
		return
	}
	helpers.WriteJSONResponse(w, http.StatusOK, "Cart emptied successfully")
}

func (s *Server) CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	userID := s.db.GetUserID(r)
	cart, err := s.db.GetCart(userID)
	if err != nil {
		helpers.HandleError(w, http.StatusNotFound, "Cart does not exist")
		return
	}

	if cart.Quantity == 0 {
		helpers.HandleError(w, http.StatusBadRequest, "Cart is empty")
		return
	}

	if err := s.db.ProcessCheckout(cart); err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, "Failed to process checkout")
		return
	}

	helpers.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Order placed"})
}

/////////////

func (s *Server) IndexVendorsHandler(w http.ResponseWriter, r *http.Request) {
	vendors, meta, err := s.db.ListVendors(r.URL.Query())
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}
	helpers.WriteJSONResponse(w, http.StatusOK, types.Response{Meta: meta, Data: vendors})
}

func (s *Server) GetVendorHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	vendor, err := s.db.GetVendorByID(id)
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}
	helpers.WriteJSONResponse(w, http.StatusOK, vendor)
}

func (s *Server) CreateVendorHandler(w http.ResponseWriter, r *http.Request) {
	var vendor types.Vendor
	if err := json.NewDecoder(r.Body).Decode(&vendor); err != nil {
		helpers.HandleError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	createdVendor, err := s.db.CreateVendor(vendor)
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}
	helpers.WriteJSONResponse(w, http.StatusCreated, createdVendor)
}

func (s *Server) UpdateVendorHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var vendor types.Vendor
	if err := json.NewDecoder(r.Body).Decode(&vendor); err != nil {
		helpers.HandleError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	updatedVendor, err := s.db.UpdateVendor(vendor, id)
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}
	helpers.WriteJSONResponse(w, http.StatusOK, updatedVendor)
}

func (s *Server) DeleteVendorHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := s.db.DeleteVendor(id)
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := map[string]string{
		"message":  "Deleted vendor successfully!",
		"vendorID": id,
	}
	helpers.WriteJSONResponse(w, http.StatusOK, resp)
}

func (s *Server) GrantAdminHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.FormValue("user_id")
	vendorID := r.FormValue("vendor_id")
	if userID == "" || vendorID == "" {
		helpers.HandleError(w, http.StatusBadRequest, "Missing required parameters")
		return
	}
	err := s.db.GrantAdmin(userID, vendorID)
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}
	helpers.WriteJSONResponse(w, http.StatusOK, "Admin granted successfully")
}

func (s *Server) RevokeAdminHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.FormValue("user_id")
	vendorID := r.FormValue("vendor_id")
	if userID == "" || vendorID == "" {
		helpers.HandleError(w, http.StatusBadRequest, "Missing required parameters")
		return
	}
	err := s.db.RevokeAdmin(userID, vendorID)
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}
	helpers.WriteJSONResponse(w, http.StatusOK, "Admin revoked successfully")
}

func (s *Server) IndexVendorAdminsHandler(w http.ResponseWriter, r *http.Request) {
	vendorID := r.PathValue("id")
	admins, err := s.db.ListVendorAdmins(vendorID)
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}
	helpers.WriteJSONResponse(w, http.StatusOK, admins)
}

///////
