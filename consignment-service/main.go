// consignment-service/main.go

package main 

import (
	// Import the generated protobuf code
	"fmt"
	"log"
	
	pb "github.com/rizzomatic/shipping-mgmt/consignment-service/proto/consignment"
	vesselProto "github.com/rizzomatic/shipping-mgmt/vessel-service/proto/vessel"
	micro "github.com/micro/go-micro"
	"golang.org/x/net/context"
)


type Repository interface {
	Create(*pb.Consignment) (*pb.Consignment, error)
	GetAll() []*pb.Consignment
}

//dummy repo now - this simulates the use of a datastore
//will be made real at a later time
type ConsignmentRepository struct {
	consignments []*pb.Consignment
}

func (repo *ConsignmentRepository) Create(consignment *pb.Consignment) (*pb.Consignment, error) {
	updated := append(repo.consignments, consignment)
	repo.consignments = updated
	return consignment, nil
}

func (repo *ConsignmentRepository) GetAll() []*pb.Consignment {
	return repo.consignments
}

//service should implement all of the methods to satisfy the service
//defined in the protobuff definition 
type service struct {
	repo Repository
	vesselClient vesselProto.VesselServiceClient 
}

// CreateConsignment and GetConsignments
//must be implemented because they are defined in the protobuf service
// arguments are handled by the gRPC server.
func (s *service) CreateConsignment(ctx context.Context, req *pb.Consignment, res *pb.Response) error {
	//Call a client instance of the vessel service using the weight fromthe consignment
	//and amount of containsers 
	vesselResponse, err := s.vesselClient.FindAvailable(context.Background(), &vesselProto.Specification{
		MaxWeight: req.Weight,
		Capacity: int32(len(req.Containers)),
		})
	log.Printf("Found vessel: %s \n", vesselResponse.Vessel.Name)
	if err != nil {
		return err
	}

	//now set the vesselId to the vesselId we got back
	//from the vessel service
	req.VesselId = vesselResponse.Vessel.Id

	//save the consignment being passed in
	consignment, err := s.repo.Create(req)

	if err != nil {
		return err
	}

	res.Created = true
	res.Consignment = consignment
	return nil

}

func (s *service) GetConsignments(ctx context.Context, req *pb.GetRequest, res *pb.Response) error {
	consignments := s.repo.GetAll()
	res.Consignments = consignments
	return nil
}


func main() {

	repo := &ConsignmentRepository{}

	// Create a new service. Optionally include some options here.
	srv := micro.NewService(

		// This name must match the package name given in your protobuf definition
		micro.Name("go.micro.srv.consignment"),
		micro.Version("latest"),
	)

	vesselClient := vesselProto.NewVesselServiceClient("go.micro.srv.vessel", srv.Client())

	// Init will parse the command line flags.
	srv.Init()

	// Register handler
	pb.RegisterShippingServiceHandler(srv.Server(), &service{repo, vesselClient})

	// Run the server
	if err := srv.Run(); err != nil {
		fmt.Println(err)
	}
}

