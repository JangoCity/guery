package Scheduler

import (
	"sort"
	"sync"
	"time"

	"github.com/xitongsys/guery/EPlan"
	"github.com/xitongsys/guery/Plan"
	"github.com/xitongsys/guery/pb"
)

const (
	MAXPN int32 = 1000
	MINPN int32 = 1
)

type Scheduler struct {
	sync.Mutex
	Topology *Topology

	Todos, Doings, Dones, Fails TaskList
	AllocatedMap                map[string]int32 //executorName:taskId
	FreeExecutors               []pb.Location

	TotalTaskNumber int64
}

func (self *Scheduler) fresh() {
	self.Lock()
	defer self.Unlock()

	self.FreeExecutors = []pb.Location{}
	for _, einfo := range self.Topology.Executors {
		name := einfo.Heartbeat.Location.Name
		if _, ok := AllocatedMap[name]; !ok && einfo.Heartbeat.Status == 0 {
			self.FreeExecutors = append(self.FreeExecutors, *einfo.Heartbeat.Location)
		}
	}
}

func (self *Scheduler) AddTask(query, catalog, schema string, priority int32) error {
	var err error
	self.Lock()
	defer self.Unlock()

	TotalTaskNumber++
	taskId = TotalTaskNumber
	task := &Task{
		TaskId:     taskId,
		TaskStatus: TODO,
		Executors:  []string{},
		Query:      query,
		Catalog:    catalog,
		Schema:     schema,
		Priority:   priority,

		CommitTime: time.Now(),
	}

	var logicalPlanTree PlanNode
	logicalPlanTree, err = Plan.CreateLogicalTree(query)
	if err == nil {
		task.LogicalPlanTree = logicalPlanTree
		task.ExecutorNumber, err = EPlan.GetEPlanExecutorNumber(task.LogicalPlanTree, 1)
		if err == nil {
			self.Todos = append(self.Todos, task)
			sort.Sort(self.Todos)
		}
	}

	if err != nil {
		task.Fails = append(task.Fails, task)
	}

	return err
}

func (self *Scheduler) RunTask() {
	self.Lock()
	defer self.Unlock()

	task := self.Todos.Top()

	if task.ExecutorNumber > len(self.FreeExecutors) {
		return
	}

	freeExecutorsNumber := int32(len(self.FreeExecutors))

	l, r := MINPN, MAXPN
	for l <= r {
		m := l + (r-l)/2
		men, _ := EPlan.GetEPlanExecutorNumber(task.LogicalPlanTree, m)
		if men > freeExecutorsNumber {
			r = m - 1
		} else {
			l = m + 1
		}
	}
	pn := r
	task.ExecutorNumber, _ = EPlan.GetEPlanExecutorNumber(task.LogicalPlanTree, pn)
	self.Todos.Pop()

	//start send to executor
	ePlanNodes := []EPlan.ENode{}
	freeExecutors := self.FreeExecutors[:task.ExecutorNumber]

	if aggNode, err = EPlan.CreateEPlan(task.logicalPlanTree, &ePlanNodes, &freeExecutors, 1); err == nil {
		for _, enode := range ePlanNodes {
			var buf bytes.Buffer
			gob.NewEncoder(&buf).Encode(enode)

			instruction := pb.Instruction{
				TaskId:                task.TaskId,
				TaskType:              int32(enode.GetNodeType()),
				Catalog:               task.Catalog,
				Schema:                task.Schema,
				EncodedEPlanNodeBytes: buf.String(),
			}
			instruction.Base64Encode()

			loc := enode.GetLocation()
			var grpcConn *grpc.ClientConn
			grpcConn, err = grpc.Dial(loc.GetURL(), grpc.WithInsecure())
			if err != nil {
				break
			}
			client := pb.NewGueryExecutorClient(grpcConn)
			if _, err = client.SendInstruction(context.Background(), &instruction); err != nil {
				grpcConn.Close()
				break
			}

			empty := pb.Empty{}
			if _, err = client.SetupWriters(context.Background(), &empty); err != nil {
				grpcConn.Close()
				break
			}
			grpcConn.Close()
		}

		if err == nil {
			for _, enode := range ePlanNodes {
				loc := enode.GetLocation()
				var grpcConn *grpc.ClientConn
				grpcConn, err = grpc.Dial(loc.GetURL(), grpc.WithInsecure())
				if err != nil {
					break
				}
				client := pb.NewGueryExecutorClient(grpcConn)
				empty := pb.Empty{}

				if _, err = client.SetupReaders(context.Background(), &empty); err != nil {
					Logger.Errorf("failed setup readers %v: %v", loc, err)
					grpcConn.Close()
					break
				}

				if _, err = client.Run(context.Background(), &empty); err != nil {
					Logger.Errorf("failed run %v: %v", loc, err)
					grpcConn.Close()
					break
				}
				grpcConn.Close()
			}
		}
	}

	if err != nil {
		task.Status = FAILED
		self.Fails = append(self.Fails, task)
	}

}
